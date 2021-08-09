# Go-RCON

Go-RCON is a client implementation of [Valve's Source RCON Protocol](https://developer.valvesoftware.com/wiki/Source_RCON_Protocol) written in Go with many extra features beyond the scope of Valve's spec.

## What is RCON?

RCON is a TCP/IP based protocol which allows commands to be sent to a game server through a remote console, which is where the term RCON comes from.

It was first used by the Source Dedicated Server to support Valve's games, but many other games implemented the idea of RCON through their own implementations. Some notable examples are Minecraft, Rust and ARK: Survival Evolved.

## How does Go-RCON work?

Go-RCON works using two TCP socket connections. One socket is for command execution, and the other is for receiving broadcasts. If the game whose RCON implementation is being used doesn't use broadcasts, then you only need to connect the main socket to send commands.

## How do I use it?

First, you should create a `ClientConfig` instance and fill in the required fields. Currently, the following fields are required: `Host`, `Port`, `Password`.

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

To listen for broadcasts, there is an additional step. You must run the function `client.ListenForBroadcasts([]string, errors)`. This function takes in a string slice containing the broadcast channels to listen to, and `errors` is a channel where errors will be passed into incase any occur. `ListenForBroadcasts` uses goroutines which is why the error channel is needed. If you really don't care about errors, you can pass in `nil`. Example:

```
errorChannel := make(chan error)

client.ListenForBroadcasts([]string{"all"}, errorChannel)

// Enter loop to check for errors occurred
for {
    select {
        case err := <- errors:
        // handle error
        break
    }
}
```

## Example

For a full example, check out examples/main.go in this repository.

# Contributing

If you have an idea on how we can make Go-RCON better, don't hesitate to reach out or open an issue or pull request!
