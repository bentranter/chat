package torbit

import (
	"fmt"
	"log"
	"net"
	"time"
)

type messageType int

const (
	defaultChannelName = "general"

	join = messageType(iota)
	leave
	text
	quit
)

type message struct {
	channel     string
	username    string
	text        string
	time        time.Time
	messageType messageType
}

func newMessage(channel, username, text string, messageType messageType) *message {
	return &message{
		channel:     channel,
		username:    username,
		text:        text,
		time:        time.Now(),
		messageType: messageType,
	}
}

type channel struct {
	name        string
	users       map[*User]bool
	activeUsers map[string]*User
	broadcastCh chan *message
}

func newChannel(channelName string, activeUsers map[string]*User) *channel {
	ch := &channel{
		name:        channelName,
		users:       make(map[*User]bool),
		activeUsers: make(map[string]*User),
		broadcastCh: make(chan *message),
	}
	ch.activeUsers = activeUsers
	go ch.broadcast()
	return ch
}

func (c *channel) join(u *User) {
	c.users[u] = true
	c.broadcastCh <- newMessage(c.name, u.name, u.name+" has joined\n", text)
	go u.conn.read()
}

func (c *channel) leave(u *User) {
	delete(c.users, u)
}

func (c *channel) broadcast() {
	for {
		msg := <-c.broadcastCh
		for u := range c.users {
			if _, ok := c.activeUsers[u.name]; !ok {
				continue
			}
			err := u.conn.write(fmt.Sprintf("(%s to %s): %s", msg.username, msg.channel, msg.text))
			if err != nil {
				println("Broadcast error: ", err.Error())
			}
		}
	}
}

type hub struct {
	logger    *log.Logger
	channels  map[string]*channel
	users     map[string]*User
	userCh    chan *User
	messageCh chan *message
}

func newHub(l *log.Logger) *hub {
	return &hub{
		logger:    l,
		channels:  make(map[string]*channel),
		users:     make(map[string]*User),
		userCh:    make(chan *User),
		messageCh: make(chan *message),
	}
}

func (h *hub) run() {
	h.channels[defaultChannelName] = newChannel(defaultChannelName, h.users)
	for {
		select {
		case user := <-h.userCh:
			h.users[user.name] = user
			h.channels[defaultChannelName].join(user)

		case message := <-h.messageCh:
			switch message.messageType {

			case join:
				h.channels[message.channel].users[h.users[message.username]] = true

			case leave:
				delete(h.channels[message.channel].users, h.users[message.username])

			case text:
				h.logger.Printf("(%s to %s): %s", message.username, message.channel, message.text)
				h.channels[message.channel].broadcastCh <- message

			case quit:
				h.logger.Printf("%s", message.text)
				h.users[message.username].conn.close() // pls panic
				delete(h.users, message.username)
				h.channels[defaultChannelName].broadcastCh <- message
			}
		}
	}
}

func (h *hub) serve(port string) error {
	server, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	h.logger.Println("Server started on", port)

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				h.logger.Println(err.Error())
			}
			newUser := createTCPUser(conn, h)
			h.userCh <- newUser
		}
	}()

	h.run()
	return nil
}

func ListenAndServe(l *log.Logger, port string) error {
	h := newHub(l)
	return h.serve(port)
}

// func (s *server) serve(port string) error {
// 	// HTTP Server/Websocket server
// 	go func() {
// 		http.HandleFunc("/", homeHandler)
// 		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
// 			newWsClientHandler(s, w, r)
// 		})
// 		http.ListenAndServe(":8000", nil)
// 	}()
// }

// // @TODO: This needs to be a template so the port/ip can be set!
// const homeHTML = `<!DOCTYPE html>
// <html>
//   <head>
//     <meta charset="utf-8"/>
//     <title>Chat</title>
//     <meta name="viewport" content="width=device-width,initial-scale=1"/>
//     <link href="https://npmcdn.com/basscss@8.0.0/css/basscss.min.css" rel="stylesheet">
//     <style>
//       html, body { font-family: "Proxima Nova", Helvetica, Arial, sans-serif }
//       .bg-blue { background-color: #07c }
//       .white { color: #fff }
//       .bold { font-weight: bold }
//     </style>
//   </head>

//   <body class="p2">
//     <h1 class="h1">Welcome to the chat room!</h1>
//     <form id="form" class="flex">
//       <input class="flex-auto px2 py1 bg-white border rounded" type="text" id="msg">
//       <input class="px2 py1 bg-blue white bold border rounded" type="submit" value="Send">
//     </form>
//     <div class="my2" id="box"></div>
//   <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
//   <script>
//   $(function() {

//     var ws = new window.WebSocket("ws://" + document.domain + ":8000/ws");
//     var $msg = $("#msg");
//     var $box = $("#box");

//     ws.onclose = function(e) {
//       $box.append("<p class='bold'>Connection closed!</p>");
//     };
//     ws.onmessage = function(e) {
//       $box.append("<p>"+e.data+"</p>");
//       increaseUnreadCount();
//     };

//     ws.onerror = function(e) {
//       $box.append("<strong>Error!</strong>")
//     };

//     $("#form").submit(function(e) {
//       e.preventDefault();
//       if (!ws) {
//           return;
//       }
//       if (!$msg.val()) {
//           return;
//       }
//       ws.send($msg.val());
//       $msg.val("");
//     });

//     document.addEventListener("visibilitychange", resetUnreadCount);

//     function increaseUnreadCount() {
//       if (document.hidden === true) {
//         var count = parseInt(document.title.match(/\d+/));
//         if (!count) {
//           document.title = "(1) Chat";
//           return;
//         }
//         document.title = "("+(count+1)+") Chat";
//       }
//     }

//     function resetUnreadCount() {
//       if (document.hidden === false) {
//         document.title = "Chat";
//       }
//     }

//   });
//   </script>
//   </body>
// </html>
// `
