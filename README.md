Torbit Go Programming Challenge
---

A very simple TCP chat app.

Usage
---

To run this, start the server:

```go
$ go run main.go
```

Then connect to it via telnet by doing

```
$ telnet
> telnet localhost 3000
```

To exit telnet, hit _Ctrl-]_ to disconnect, then type _quit_.

Options
---

The server will accept the following flags:

```
  -http string
      http service address (default "8000")
  -ip string
      ip service address (default "localhost")
  -log string
      log file location (default "stdout")
  -tcp string
      tcp service address (default "3000")
```
