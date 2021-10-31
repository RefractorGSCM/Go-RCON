package rcon

import (
	"bufio"
	"github.com/pkg/errors"
	"github.com/refractorgscm/rcon/errs"
	"github.com/refractorgscm/rcon/packet"
	"strings"
	"time"
)

func (c *Client) sendPacket(p packet.Packet) error {
	out, err := p.Build()
	if err != nil {
		return errors.Wrap(err, "could not build packet")
	}

	if err := c.write(out); err != nil {
		return errors.Wrap(err, "could not send authentication packet")
	}

	return nil
}

func (c *Client) readPacket() (packet.Packet, error) {
	if c.conn == nil {
		return nil, errs.ErrNotConnected
	}

	if err := c.conn.SetDeadline(time.Time{}); err != nil {
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			return nil, errs.ErrNotConnected
		}

		return nil, errors.Wrap(err, "could not set connection deadline")
	}

	reader := bufio.NewReader(c.conn)

	res, err := packet.DecodeClientPacket(c.EndianMode, reader)
	if err != nil {
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			return nil, errs.ErrNotConnected
		}

		return nil, errors.Wrap(err, "could not read packet")
	}

	return res, nil
}

func (c *Client) readPacketTimeout() (packet.Packet, error) {
	if c.conn == nil {
		return nil, errs.ErrNotConnected
	}

	if err := c.conn.SetDeadline(time.Now().Add(c.ConnTimeout)); err != nil {
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			return nil, errs.ErrNotConnected
		}

		return nil, errors.Wrap(err, "could not set connection deadline")
	}

	reader := bufio.NewReader(c.conn)

	res, err := packet.DecodeClientPacket(c.EndianMode, reader)
	if err != nil {
		if strings.HasSuffix(err.Error(), "use of closed network connection") {
			return nil, errs.ErrNotConnected
		}

		return nil, errors.Wrap(err, "could not read packet")
	}

	return res, nil
}

func (c *Client) write(data []byte) error {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	if _, err := c.conn.Write(data); err != nil {
		return err
	}

	return nil
}
