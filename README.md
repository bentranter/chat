Chat
---

A TCP and WebSocket powered chatroom.

#####Features

1. Access over TCP via telnet (or over TCP with TLS via openssl s_client)
2. Send receive messages
3. Multiple chat rooms
4. Support for commands, including muting users and direct messaging users.
5. Reading from a config file, which can be overridden via flags.
6. Support for access over websockets (but no JavaScript client :( )
7. Proof-of-concept support for a REST API.

Usage
---

To run this, start the server:

```go
$ go run main.go
```

Or build in via `go build`.

If you wish to pass in a config file, that file **must** be in the same directory as you start the server from. The config file currently must been named `config.toml`, and requires valid TOML syntax.

The following command line flags are also accepted:

```
  -http string
      http port (default "8000")
  -https string
      https port (default "8001")
  -ip string
      ip address (default "localhost")
  -log string
      log filename (default "stdout")
  -tcp string
      tcp port (default "3000")
  -tcps string
      secure tcp port (default "3001")
```

If a filename is passed for the logfile, a multiwriter will be used to write to both that file _and_ stdout.

Clients
---

Chat works with four(ish) different clients:

#####TCP (via Telnet)

Connect to it via telnet by doing

```
$ telnet localhost <port>
```

Where port is the address that the insecure TCP server starts on (default is 3000). To exit telnet, hit _Ctrl-]_ to disconnect, then type _quit_.

#####OpenSSL's Telnet Thing

Connect by doing:

```
$ openssl s_client -connect <ipAddr>:<port>
```

Where ipAddr is the IP address of the server (default is localhost), and the port is the secure TCP server (default is 3001). To exit this thing just hit Ctrl-c. Also, don't type a `B` ever, or it will try to send a heartbeat and crash :P

#####API

The API exists only as a proof of concept. It will always respond with what you sent, whether it sent successfully or not.

```bash
curl -X POST -H "Content-Type: application/json" -d '{
    "Channel": "general",
    "Username": "<username>",
    "Text": "Hello from the API!",
    "MessageType": 6
}' "<protocol>://<ipAddr>:<port>/messages"
```

where, like above, ipAddr is the IP address (default: localhost), and the port is that of the HTTP server (default: 8000). The protocol here can either be HTTP or HTTPS, although the port for HTTPS will be different (default is 8001).

#####Websockets

Like the API, the websocket implementation exists as a proof of concept. You can connect by sending a `POST` request with your desired username as JSON to `/ws`. It communicates with the server by sending `message`s encoded as JSON. Requests can be sent to the HTTP or HTTPS server, with values reflecting the ones listed above in the API section.
