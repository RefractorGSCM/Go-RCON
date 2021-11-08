package main

import (
	"fmt"
	"github.com/refractorgscm/rcon"
	"github.com/refractorgscm/rcon/presets"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	client := rcon.NewClient(&rcon.Config{
		Host:     "127.0.0.1",
		Port:     7779,
		Password: "RconPassword",
		BroadcastHandler: func(msg string) {
			fmt.Println("RECEIVED BROADCAST", msg)
		},
		RestrictedPacketIDs: presets.MordhauRestrictedPacketIDs,
		BroadcastChecker:    presets.MordhauBroadcastChecker,
		DisconnectHandler: func(err error, expected bool) {
			if !expected {
				log.Println("An unexpected disconnection has occurred. Error:", err)
			} else {
				log.Println("An expected disconnection has occurred.")
			}
		},
	}, &presets.DebugLogger{})

	if err := client.Connect(); err != nil {
		log.Fatalf("Could not connect. Error: %v\n", err)
	}

	res, err := client.ExecCommand("listen chat")
	if err != nil {
		log.Fatalf("Could not execute command. Error: %v\n", err)
	}
	fmt.Println(res)

	// Cleanup on CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		log.Println("Shutting down...")
		_ = client.Close()
	}()

	client.WaitGroup().Wait()
}
