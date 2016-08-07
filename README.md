Chat
---

A TCP and WebSocket powered chatroom.

Usage
---

To run this, start the server:

```go
$ go run main.go
```

Or build in via `go build`.

If you wish to pass in a config file, that file must be in the same directory as you start the server from. The config file currently must been named `config.toml`, and requires valid TOML syntax.

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

To exit telnet, hit _Ctrl-]_ to disconnect, then type _quit_.

#####OpenSSL's Telnet Thing

Connect by doing:

```
$ openssl s_client -connect <ipAddr>:<port>
```

To exit this thing just hit Ctrl-c. Also, don't type a `B` ever, or it will try to send heartbeat and crash :P

#####Web

_TODO_.
