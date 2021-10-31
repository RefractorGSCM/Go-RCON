package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"github.com/refractorgscm/rcon/endian"
	"io"
	"math"
)

var nextClientPacketID int32 = 0

type ClientPacket struct {
	mode  endian.Mode
	pType PacketType
	body  []byte
	id    int32
}

func idInArr(arr []int32, id int32) bool {
	for _, v := range arr {
		if v == id {
			return true
		}
	}

	return false
}

func getNextID(restrictedIDs []int32) int32 {
	if nextClientPacketID+1 == math.MaxInt32 {
		nextClientPacketID = 1
	} else {
		nextClientPacketID++
	}

	// Check if the current nextClientPacketID is a restricted id and increment it until it no longer is
	for idInArr(restrictedIDs, nextClientPacketID) {
		if nextClientPacketID+1 == math.MaxInt32 {
			nextClientPacketID = 1
		} else {
			nextClientPacketID++
		}
	}

	return nextClientPacketID
}

func NewClientPacket(mode endian.Mode, pType PacketType, body string, restrictedIDs []int32) Packet {
	nextClientPacketID = getNextID(restrictedIDs)

	p := &ClientPacket{
		mode:  mode,
		pType: pType,
		body:  []byte(body),
		id:    nextClientPacketID,
	}

	if len(body) == 0 {
		p.body = []byte{}
	}

	return p
}

const int32Bytes = 4
const endPadBytes = 1

func (p *ClientPacket) Size() int32 {
	return int32Bytes + int32Bytes + int32(len(p.Body())) + endPadBytes
}

func (p *ClientPacket) ID() int32 {
	return p.id
}

func (p *ClientPacket) Type() PacketType {
	return p.pType
}

func (p *ClientPacket) Body() []byte {
	return append(p.body, byte('\x00'))
}

func (p *ClientPacket) Build() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})

	order := p.mode

	if err := binary.Write(buffer, order, p.Size()); err != nil {
		return nil, errors.Wrap(err, "could not write packet size")
	}

	if err := binary.Write(buffer, order, p.ID()); err != nil {
		return nil, errors.Wrap(err, "could not write packet size")
	}

	if err := binary.Write(buffer, order, p.Type()); err != nil {
		return nil, errors.Wrap(err, "could not write packet size")
	}

	if err := binary.Write(buffer, order, p.Body()); err != nil {
		return nil, errors.Wrap(err, "could not write packet size")
	}

	if err := binary.Write(buffer, order, byte('\x00')); err != nil {
		return nil, errors.Wrap(err, "could not write packet size")
	}

	return buffer.Bytes(), nil
}

var malformedPacketErr = fmt.Errorf("malformed packet")

func DecodeClientPacket(mode endian.Mode, reader io.Reader) (*ClientPacket, error) {
	var size int32
	var id int32
	var pType int32

	// Read size
	if err := binary.Read(reader, mode, &size); err != nil {
		return nil, err
	}

	// Read ID
	if err := binary.Read(reader, mode, &id); err != nil {
		return nil, err
	}

	// Read type
	if err := binary.Read(reader, mode, &pType); err != nil {
		return nil, err
	}

	// Read body
	bodyLen := size - 4 - 4 // size - id bytes - type bytes
	body := make([]byte, bodyLen)

	_, err := io.ReadFull(reader, body)
	if err != nil {
		return nil, err
	}

	// Trim unneeded bytes from body
	body = bytes.Trim(body, "\x00")
	body = bytes.Trim(body, "\n")

	// Construct and return client packet
	return &ClientPacket{
		mode:  mode,
		pType: PacketType(pType),
		body:  body,
		id:    id,
	}, nil
}
