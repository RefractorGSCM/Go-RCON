![GitHub](https://img.shields.io/github/license/RefractorGSCM/RCON?style=flat-square)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/RefractorGSCM/RCON?style=flat-square)
![GitHub contributors](https://img.shields.io/github/contributors/RefractorGSCM/RCON?style=flat-square)

# Go-RCON

*This repository is simply named RCON, but will be referred to as Go-RCON for clarity.*

Go-RCON is a client implementation of [Valve's Source RCON Protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol) written in Go with some extra features beyond the scope of Valve's spec.
This client was designed with cross-game compatability in mind.

> Version 1.0.0 released!

## What is RCON?

RCON is a TCP/IP based protocol which allows commands to be sent to a game server through a remote console, which is where the term RCON comes from.

It was first used by the Source Dedicated Server to support Valve's games, but many other games implemented the idea of RCON through their own implementations. Some notable examples are Minecraft, Rust and ARK: Survival Evolved.

## How does Go-RCON work?

Go-RCON works by opening a TCP connection to the provided game server and authenticating with the provided password.

If authentication was successful, it will begin waiting for commands and listening for broadcast messages.

Go-RCON makes use of queues to handle delivery and separation of messages. Consider the below scenario:

1. The client is authenticated
2. ExecCommand is called to execute a command on the server.
   1. A packet containing the command data is created.
   2. The packet is queued on the write queue.
3. The internal writer routine writes the data the TCP connection (sends the command to the server).
   1. A "mailbox" is created for the command with the sent packet's ID.
4. The internal reader routine reads the response from the server and adds it to the correct mailbox.
5. The mailbox is read and the command's response is returned.

The internal reader routine uses the provided `BroadcastChecker` function to determine if a received packet
is a broadcast packet. If it is, it is sent to the provided `BroadcastHandler` and does not get forwarded to
any mailboxes.

### What are broadcast messages?

Broadcast messages are any data which was proactively sent by the game server without being directly requested by an RCON client.

This could for example be chat messages, player join/quit notifications, and any other information the server
proactively sends over.

Please note that many games do not support broadcast messages. The broadcast message system was
designed with the game Mordhau in mind.

## How do I use it?

First, you should create a `Config` instance and fill in the required fields.
Currently, the following fields are required: `Host`, `Port`, `Password`.

```
clientConfig := &rcon.ClientConfig{
	Host:     host,
	Port:     port,
	Password: password,
	// etc
}

client := rcon.NewClient(clientConfig)
```

### Connecting to the RCON server

Once your client is configured to your requirements, connect the client to your RCON server using `client.Connect()`. Example:

```
if err := client.Connect(); err != nil {
    // handle error
}
```

### Executing commands

Once the client is connected to your RCON server, you can start sending commands using `client.ExecCommand(string)`. Example:

```
response, err := client.ExecCommand("PlayerList")
if err != nil {
    // handle error
}

// do something with response
```

### Listening for broadcasts

Broadcasts are listened for automatically, however you need to instruct your RCON client how to determine if a packet is
a broadcast.

You do this by creating a `BroadcastMessageChecker` which is a function with the following signature:

```
func (p packet.Packet) bool
```

For an example, check out the Mordhau broadcast checker preset in `presets/broadcast_checkers.go`.

Once that's done, you should set a broadcast handler function. This function will be called whenever a broadcast message
is received. It should have the following signature:

```
func (message string)
```

### Handling Disconnects

In the case of a disconnection, the provided `DisconnectHandler` function is called.
This is the `DisconnectHandler` signature:

```
func (err error, expected bool)
```

If the disconnect was expected, expected will be true and error will be nil. Otherwise, expected will be false
and the error causing the disconnect will be set in err.

An expected disconnect only happens if you call `client.Close()`.

### Reconnecting After a Disconnect

Go-RCON has no built in reconnect routine. This is because different applications will require
different methods of reconnect as well as varying levels of control over the reconnection. Because of this it
doesn't make much sense to add a built-in reconnect routine since it would likely not see much use.

If you need automatic reconnection, I would suggest detecting the disconnect using a `DisconnectHandler` (set in the
client config) and then kicking off your own reconnect routine.

## Example

For a full example, check out examples/main.go in this repository.

# Contributing

Contributions are welcome! If you have an idea to make Go-RCON better, bug fixes or any other changes feel free to open
an issue and a corresponding pull request.
