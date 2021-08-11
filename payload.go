package rcon

import (
	"fmt"
	"net"
	"regexp"
)

const (
	packetIDAuthFailed = -1
	payloadIDBytes     = 4
	payloadTypeBytes   = 4
	payloadNullBytes   = 2
	payloadMaxSize     = 2048
)

var (
	currentPayloadID int32 = 0
)

type payload struct {
	ID   int32
	Type int32
	Body []byte

	// NonBroadcastPatterns is where known non broadcast patterns can be added to the ignore list so that payloads
	// which aren't broadcasts are detectable.
	NonBroadcastPatterns []*regexp.Regexp
}

func newPayload(payloadType int, body []byte, nonBroadcastPatterns []*regexp.Regexp) *payload {
	currentPayloadID++

	if nonBroadcastPatterns == nil {
		nonBroadcastPatterns = []*regexp.Regexp{}
	}

	return &payload{
		ID:                   currentPayloadID,
		Type:                 int32(payloadType),
		Body:                 body,
		NonBroadcastPatterns: nonBroadcastPatterns,
	}
}

func (p *payload) getSize() int32 {
	return int32(len(p.Body) + (payloadIDBytes + payloadTypeBytes + payloadNullBytes))
}

func (p *payload) isNotBroadcast() bool {
	for _, pattern := range p.NonBroadcastPatterns {
		// If payload body matches a known non-broadcast pattern, we can safely
		// assume that it's not a broadcast so we return true.
		if pattern.MatchString(string(p.Body)) {
			return true
		}
	}

	// If none of the known non-broadcast patterns were matches, return false.
	return false
}

func sendPayload(conn net.Conn, request *payload) (*payload, error) {
	packet, err := buildPacketFromPayload(request)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(packet)
	if err != nil {
		return nil, err
	}

	response, err := buildPayloadFromPacket(conn)
	if err != nil {
		return nil, err
	}

	if response.ID == packetIDAuthFailed {
		return nil, fmt.Errorf("authentication failed")
	}

	return response, nil
}
