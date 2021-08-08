package rcon

import (
	"fmt"
	"net"
	"testing"
)

// setupTestServer spins up a very basic TCP server to emulate an RCON server.
// This server is for testing purposes only.
func testServerSetup(t *testing.T, ready chan bool) error {
	listener, err := net.Listen("tcp", "localhost:7891")
	if err != nil {
		t.Errorf("Could not start test TCP server. Error: %v", err)
	}
	defer listener.Close()

	ready <- true

	conn, err := listener.Accept()
	if err != nil {
		t.Errorf("Could not accept TCP connection on test TCP server. Error: %v", err)
	}

	err = testServerhandleRequest(conn)
	if err != nil {
		fmt.Println(err)
	}

	conn.Close()

	return nil
}

func testServerhandleRequest(conn net.Conn) error {
	buffer := make([]byte, 128)

	size, err := conn.Read(buffer)
	if err != nil {
		return err
	}

	// Output message
	fmt.Printf("Read %d bytes: [", size)
	for i := 0; i < size; i++ {
		fmt.Printf("%v ", buffer[i])
	}
	fmt.Print("]\n")

	// Send back basic known data
	payload := newPayload(serverDataExecCommand, []byte("Hello!"), nil)
	packet, err := buildPacketFromPayload(payload)
	if err != nil {
		return err
	}

	_, err = conn.Write(packet)
	if err != nil {
		return err
	}

	conn.Close()

	return nil
}
