package rcon

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/refractorgscm/rcon/endian"
	"github.com/refractorgscm/rcon/errs"
	"github.com/refractorgscm/rcon/packet"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	*Config
	conn     *net.TCPConn
	connLock sync.Mutex
	log      Logger

	terminate  chan uint8
	waitGroup  *sync.WaitGroup
	wqLock     sync.Mutex
	rqLock     sync.Mutex
	wgLock     sync.Mutex
	writeQueue chan packet.Packet
	readQueue  map[int32]chan packet.Packet
}

type BroadcastHandler func(string)
type BroadcastMessageChecker func(p packet.Packet) bool
type DisconnectHandler func(error, bool)

type Config struct {
	Host     string
	Port     uint16
	Password string

	// ConnTimeout is the timeout for TCP connection read/write operations with a deadline.
	ConnTimeout time.Duration

	// QueueWriteTimeout is the timeout for writing to the internal packet queues. Higher values can cause delays if
	// unexpected packets are received.
	//
	// Default: 250ms
	QueueWriteTimeout time.Duration

	// QueueReadTimeout is the timeout for reading from the internal packet queues.
	//
	// Default: 2s
	QueueReadTimeout time.Duration

	// EndianMode represents the byte order being used by whatever game you're using this library with. Valve games
	// typically use little endian, but other games may use big endian. You can switch this as needed.
	EndianMode endian.Mode

	// BroadcastHandler is a function which will be called with a message whenever a broadcast message is received.
	BroadcastHandler BroadcastHandler

	// BroadcastChecker is a function which should be implemented. It is used to check if a packet is a broadcast.
	// If BroadcastChecker returns true, the packet will be treated as a broadcast.
	BroadcastChecker BroadcastMessageChecker

	// RestrictedPacketIDs is a slice of int32s which cannot be used as packet IDs. Some games use certain packet IDs to
	// denote a special response or message. For example, Mordhau uses these packet IDs to denote broadcast messages.
	//
	// The special packet IDs of whatever game you're using this library with should be put within this slice to ensure
	// that the received and sent data is as you'd expect and to avoid potential client/server confusion.
	RestrictedPacketIDs []int32

	// DisconnectHandler is a function which will be called when the client gets disconnected.
	DisconnectHandler DisconnectHandler
}

const DefaultTimeout = time.Second * 2

func NewClient(config *Config, logger Logger) *Client {
	c := &Client{
		Config:     config,
		log:        &DefaultLogger{},
		waitGroup:  &sync.WaitGroup{},
		terminate:  make(chan uint8),
		writeQueue: make(chan packet.Packet),
		readQueue:  map[int32]chan packet.Packet{},
	}

	if logger != nil {
		c.log = logger
	}

	if c.EndianMode == nil {
		c.EndianMode = endian.Little
	}

	if c.ConnTimeout <= 0 {
		c.ConnTimeout = DefaultTimeout
	}

	if c.BroadcastChecker == nil {
		c.BroadcastChecker = func(p packet.Packet) bool {
			return false
		}
	}

	if c.QueueWriteTimeout <= 0 {
		c.QueueWriteTimeout = time.Millisecond * 250
	}

	if c.QueueReadTimeout <= 0 {
		c.QueueReadTimeout = time.Second * 2
	}

	return c
}

func (c *Client) SetBroadcastHandler(handler BroadcastHandler) {
	c.BroadcastHandler = handler
}

func (c *Client) SetDisconnectHandler(handler DisconnectHandler) {
	c.DisconnectHandler = handler
}

func (c *Client) SetBroadcastChecker(checker BroadcastMessageChecker) {
	c.BroadcastChecker = checker
}

func (c *Client) SetRestrictedPacketIDs(restrictedIDs []int32) {
	c.RestrictedPacketIDs = restrictedIDs
}

func (c *Client) Connect() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), c.ConnTimeout)
	if err != nil {
		return errors.Wrap(err, "tcp dial failure")
	}
	c.log.Debug("Dial successful, connection established.")

	var ok bool
	c.conn, ok = conn.(*net.TCPConn)
	if !ok {
		return errors.Wrap(err, "tcp dial failure")
	}

	if err := c.conn.SetDeadline(time.Now().Add(c.ConnTimeout)); err != nil {
		return errors.Wrap(err, "could not set tcp connection deadline")
	}

	if err := c.authenticate(); err != nil {
		c.log.Debug("Authentication failed", err)
		return err
	}

	c.log.Debug("Starting writer routine")
	go func() {
		c.wgLock.Lock()
		c.waitGroup.Add(1)
		c.wgLock.Unlock()
		c.startWriter()
	}()

	c.log.Debug("Starting reader routine")
	go func() {
		c.wgLock.Lock()
		c.waitGroup.Add(1)
		c.wgLock.Unlock()
		c.startReader()
	}()

	return nil
}

func (c *Client) startWriter() {
	defer func() {
		c.wgLock.Lock()
		c.waitGroup.Done()
		c.wgLock.Unlock()
		c.log.Debug("Writer routine terminated")
	}()

	for {
		select {
		case p := <-c.writeQueue:
			if err := c.sendPacket(p); err != nil {
				c.log.Debug("Could not write packet. Error: ", err)
			}
			break
		case <-c.terminate:
			c.log.Debug("Writer routine received termination signal")
			return
		}
	}
}

func (c *Client) startReader() {
	defer func() {
		c.wgLock.Lock()
		c.waitGroup.Done()
		c.wgLock.Unlock()
		c.log.Debug("Reader routine terminated")
	}()

	terminate := false

	readChan := make(chan packet.Packet)

	// Start select routine
	go func() {
		for {
			// Add packet to mailbox
			select {
			case p := <-readChan:
				c.readQueue[p.ID()] <- p
				c.log.Debug("Packet added to mailbox ID: ", p.ID())
				break
			case <-c.terminate:
				terminate = true
				c.log.Debug("Reader routine received termination signal")
				return
			}
		}
	}()

	for {
		// Break out of the loop if we're meant to terminate this routine.
		// We can be sure that terminate will be reached beyond the blocking readPacket call because the connection
		// was closed before we received the termination signal, so the blocking readPacket call will error out and
		// not block the termination instruction.
		if terminate {
			break
		}

		p, err := c.readPacket()
		if err != nil {
			switch errors.Cause(err) {
			case errs.ErrNotConnected:
				break
			case io.EOF:
				c.log.Error("Disconnected by the server. Error: ", err)
				c.disconnect(err)
				break
			case io.ErrClosedPipe:
				c.disconnect(err)
				c.log.Error("Attempted to read from a closed pipe. Error: ", err)
				break
			default:
				c.log.Debug("Reader error: ", err)
			}

			continue
		}

		packetID := p.ID()

		// Check if this packet is a broadcast message
		if c.BroadcastChecker(p) {
			c.log.Debug("Packet ", packetID, " is a broadcast message")

			// If this packet is a broadcast, notify broadcast listener and jump to next read.
			if c.BroadcastHandler != nil {
				newBody := p.Body()
				newBody = newBody[:len(newBody)-1] // strip null terminator

				c.BroadcastHandler(string(newBody))
			}

			continue
		} else {
			c.log.Debug("Packet ", packetID, " was not a broadcast", p.Type(), string(p.Body()))

			// Put packet on the read channel if it's not a broadcast
			select {
			case readChan <- p:
				break
			case <-time.After(c.QueueWriteTimeout):
				c.log.Debug("Packet ", packetID, " was unexpected (no open mailbox)")
				break
			}
		}
	}
}

func (c *Client) Close() error {
	c.log.Debug("Close called")

	if c.conn == nil {
		return errs.ErrNotConnected
	}

	c.disconnect(nil)

	return nil
}

func (c *Client) disconnect(err error) {
	// Closing the termination channel makes all routines return
	close(c.terminate)

	_ = c.conn.Close()
	c.conn = nil

	if c.DisconnectHandler != nil {
		c.DisconnectHandler(err, err == nil)
	}
}

func (c *Client) authenticate() error {
	p := c.newClientPacket(packet.TypeAuth, c.Password)

	if err := c.sendPacket(p); err != nil {
		return errors.Wrap(err, "could not send packet")
	}

	res, err := c.readPacketTimeout()
	if err != nil {
		return errors.Wrap(err, "could not get auth response")
	}

	if res.Type() != packet.TypeAuthRes {
		return errors.Wrap(err, "packet was not of the type auth response")
	}

	if res.ID() == packet.AuthFailedID {
		return errors.Wrap(errs.ErrAuthentication, "authentication failed")
	}

	c.log.Debug("Authenticated successfully")

	return nil
}

func (c *Client) WaitGroup() *sync.WaitGroup {
	return c.waitGroup
}

func (c *Client) ExecCommand(command string) (string, error) {
	p := c.newClientPacket(packet.TypeCommand, command)

	c.log.Debug("Executing command: ", command)

	if err := c.enqueuePacket(p, true); err != nil {
		return "", errors.Wrap(err, "could not enqueue command packet")
	}

	res, err := c.getResponse(p.ID())
	if err != nil {
		return "", errors.Wrap(err, "could not get command response")
	}

	// Trim off null terminator
	body := res.Body()
	body = body[:len(body)-1]

	return string(body), nil
}

func (c *Client) ExecCommandNoResponse(command string) error {
	p := c.newClientPacket(packet.TypeCommand, command)

	c.log.Debug("Executing command (expecting no response): ", command)

	if err := c.enqueuePacket(p, false); err != nil {
		return errors.Wrap(err, "could not enqueue command packet")
	}

	return nil
}

func (c *Client) enqueuePacket(p packet.Packet, createMailbox bool) error {
	// We use c.QueueWriteTimeout to set a timeout for packet queuing. If something happens and the packet cannot be put onto the
	// queue within the set timeout, an error is returned.
	select {
	case c.writeQueue <- p:
		c.log.Debug("Packet queued", " ID: ", p.ID())

		if createMailbox {
			// Create a mailbox for this packet. A mailbox is simply a channel which responses will be put on.
			c.readQueue[p.ID()] = make(chan packet.Packet)
		}

		return nil
	case <-time.After(c.QueueWriteTimeout):
		c.log.Debug("Packet queue timed out", " ID: ", p.ID())
		return errors.Wrap(errs.ErrQueueTimeout, "packet queue operation timed out")
	}
}

func (c *Client) getResponse(packetID int32) (packet.Packet, error) {
	defer func() {
		// When read operation is complete, delete packet mailbox.
		c.rqLock.Lock()
		close(c.readQueue[packetID])
		delete(c.readQueue, packetID)
		c.rqLock.Unlock()
	}()

	// We use c.QueueReadTimeout to set a timeout for response fetching. If something happens and no response can be pulled from
	// the mailbox with the provided packet ID within the set timeout period, an error is returned.
	select {
	case p := <-c.readQueue[packetID]:
		c.log.Debug("Packet removed from mailbox ID: ", packetID)
		return p, nil
	case <-time.After(c.QueueReadTimeout):
		return nil, errors.Wrap(errs.ErrReadTimeout, "mailbox read operation timed out")
	}
}

// newClientPacket is a wrapper function for packet.NewClientPacket. It makes creating packets a bit easier by automatically
// populating client-specific fields so that this doesn't need to be done manually.
func (c *Client) newClientPacket(pType packet.PacketType, body string) packet.Packet {
	return packet.NewClientPacket(c.EndianMode, pType, body, c.RestrictedPacketIDs)
}
