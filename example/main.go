package main

import (
	"fmt"
	"github.com/refractorgscm/rcon"
	"log"
	"regexp"
	"time"
)

var (
	host            = "localhost"
	port     uint16 = 7778
	password        = "RconPassword"
)

func main() {
	clientConfig := &rcon.ClientConfig{
		Host:                     host,
		Port:                     port,
		Password:                 password,
		EnableBroadcasts:         true,
		BroadcastHandler:         broadcastHandler,
		DisconnectHandler:        disconnectHandler,
		SendHeartbeatCommand:     true,
		HeartbeatCommandInterval: time.Second * 5,
		NonBroadcastPatterns: []*regexp.Regexp{
			regexp.MustCompile("^Keeping client alive for another [0-9]+ seconds$"),
		},
	}

	client := rcon.NewClient(clientConfig)

	// Connect the main socket to the RCON server
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// Optional but highly recommended: create an error channel to receive errors from
	// the ListenForBroadcasts goroutine.
	errors := make(chan error)

	// Connect broadcast socket to the RCON server and start listening for broadcasts
	client.ListenForBroadcasts([]string{"listen allon"}, nil)

	_, _ = client.ExecCommand("Alive")

	// Disconnect after 20 seconds
	go func() {
		time.Sleep(time.Second * 20)

		if err := client.Disconnect(); err != nil {
			fmt.Printf("Disconnect error: %v\n", err)
		}
	}()

	// Enter infinite loop to keep the program running. You wouldn't want to do this in practice.
	// Normally you would likely have a webserver or some other listening code you're running this
	// alongside which would keep the process running for you.
	for {
		select {
		case err := <-errors:
			log.Fatalf("ListenForBroadcasts error: %v", err)
		default:
			break
		}

		// Run basic command on the main RCON socket for demo purposes.
		//response, err := client.ExecCommand("PlayerList")
		//if err != nil {
		//	log.Fatal(err)
		//}

		// fmt.Println("Main Socket Response:", response)

		time.Sleep(1 * time.Second)
	}
}

func broadcastHandler(broadcast string) {
	fmt.Println("Received broadcast:", broadcast)
}

func disconnectHandler(err error, expected bool) {
	if !expected {
		fmt.Printf("An unexpected disconnect occurred. Error: %v\n", err)
	} else {
		fmt.Println("An expected disconnect occurred. All OK.")
	}
}
