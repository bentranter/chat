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

Approach
---

After trying a few different approaches, I decided to go with an approach where clients and the server (or hub, as I've called it in the source) communicate by sending each other messages over a channel. There are different message types that can be sent, such as `TEXT`, `JOIN`, `DM`, etc that represent different actions a client an perform. Messages can have four fields: a channel, a username, some text, and a message type.

The advantage of this is that adding different types of clients is easy once you have a standard set of messages a client can send. This is made easier in the case of writing Go code because a client only needs to have it's own name, and implement read, write, and close methods for sending and receiving commands, and closing their connection.

For example, in the TCP implementation, the use can type commands that will create messages to send to the server. For example, typing plaintext and hitting enter will just send that text, but typing `/newroom random` would create a message with the channel name "random", the message type "create", the username of whoever issued that command, and the text just a message to broadcast to that user letting them know if channel creation failed or succeeded.

In a websocket implementation, messages are sent and received as JSON. It's up to (for example) the JavaScript client to update the UI and send new messages based on the message it sees from the server. For the sake of example, imagine Slack: a user types a message into the textbox in the "general" channel, and hits enter. The JavaScript client would need to create and send a message of type "text" to the channel "general" with the username of whoever typed the message, and with the text set to what they typed. Upon response, the client would need to decide what to do based on the response, if it was successful, they could update the UI to show that message in the chat for all connected clients, or on error, let that client know their message failed to send.

For the API, the API could implement the same interface used for the clients, but as a default "API" user. It'd be a lot like the websocket implementation, relying on JSON to encode the messages.

Limitations & Known Bugs
---

* There are no tests.

I would typically _never_ write anything without tests, but since I hadn't settled on the design of the clients and server for a while, I delayed writing tests until I figured out the exact design to avoid re-writing tests multiple times.

* Reconnecting to the server with the same username causes a write to closed error

It'll try to write to the old connection _and_ the new connection, resulting in that error.

* The `ipAddr` flag isn't respected.

I just didn't have the time to implement it, even though it's pretty straightforward. It'll always serve on `localhost`.

* Variable names and structure of logged/broadcasted messages are inconsistent.

For example, some functions will accept something like, `(m *message)`, while others would accept `(message *message)`. I would avoid doing this in an _actual_ codebase, but since this is a prototype I didn't bother to go through and fix it.
