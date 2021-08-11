package rcon

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	type args struct {
		config *ClientConfig
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "TestNewClient-1",
			args: args{
				config: &ClientConfig{
					Host:     "127.0.0.1",
					Port:     7778,
					Password: "rconPassword",
				},
			},
			want: &Client{
				address:  fmt.Sprintf("%s:%d", "127.0.0.1", 7778),
				password: "rconPassword",
			},
		},
		{
			name: "TestNewClient-2",
			args: args{
				config: &ClientConfig{
					Host:     "localhost",
					Port:     22875,
					Password: "rconPassword",
				},
			},
			want: &Client{
				address:  fmt.Sprintf("%s:%d", "localhost", 22875),
				password: "rconPassword",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.args.config)

			if got.address != tt.want.address || got.password != tt.want.password {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Connect(t *testing.T) {
	// Spin up basic TCP server for testing purposes
	ready := make(chan bool)
	go testServerSetup(t, ready)
	<-ready

	type fields struct {
		address          string
		password         string
		broadcastHandler BroadcastHandlerFunc
		mainConn         *net.TCPConn
		broadcastConn    *net.TCPConn
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestClient_Connect-1",
			fields: fields{
				address:  "127.0.0.1:7891",
				password: "rconPassword",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				address:       tt.fields.address,
				password:      tt.fields.password,
				mainConn:      tt.fields.mainConn,
				broadcastConn: tt.fields.broadcastConn,
				config:        &ClientConfig{},
			}
			if err := c.Connect(); (err != nil) != tt.wantErr {
				t.Errorf("Client.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ExecCommand(t *testing.T) {
	// Spin up basic TCP server for testing purposes
	ready := make(chan bool)
	go testServerSetup(t, ready)
	<-ready

	// Open TCP connection to test server
	dialer := net.Dialer{Timeout: time.Second * 10}
	testConn, err := dialer.Dial("tcp", "localhost:7891")
	if err != nil {
		t.Errorf("Could not open connection to test server. Error: %v", err)
	}

	type fields struct {
		address          string
		password         string
		broadcastHandler BroadcastHandlerFunc
		mainConn         *net.TCPConn
		broadcastConn    *net.TCPConn
	}
	type args struct {
		command string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestClient_ExecCommand-1",
			fields: fields{
				address:  "localhost:7891",
				password: "rconPassword",
				mainConn: testConn.(*net.TCPConn),
			},
			args: args{
				command: "Hello",
			},
			want:    "Hello!",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				address:       tt.fields.address,
				password:      tt.fields.password,
				mainConn:      tt.fields.mainConn,
				broadcastConn: tt.fields.broadcastConn,
				config:        &ClientConfig{},
			}
			got, err := c.ExecCommand(tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ExecCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Client.ExecCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
