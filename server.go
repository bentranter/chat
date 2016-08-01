package torbit

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

const chatHelp = `(chatbot): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /name    set your name               (example: /name Ben)
  /id      view your user id           (example: /id)

`

// @TODO: This needs to be a template so the port/ip can be set!
const homeHTML = `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8"/>
    <title>Chat</title>
    <meta name="viewport" content="width=device-width,initial-scale=1"/>
    <link href="https://npmcdn.com/basscss@8.0.0/css/basscss.min.css" rel="stylesheet">
    <style>
      html, body { font-family: "Proxima Nova", Helvetica, Arial, sans-serif }
      .bg-blue { background-color: #07c }
      .white { color: #fff }
      .bold { font-weight: bold }
    </style>
  </head>

  <body class="p2">
    <h1 class="h1">Welcome to the chat room!</h1>
    <form id="form" class="flex">
      <input class="flex-auto px2 py1 bg-white border rounded" type="text" id="msg">
      <input class="px2 py1 bg-blue white bold border rounded" type="submit" value="Send">
    </form>
    <div class="my2" id="box"></div>
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
  <script>
  $(function() {

    var ws = new window.WebSocket("ws://" + document.domain + ":8000/ws");
    var $msg = $("#msg");
    var $box = $("#box");

    ws.onclose = function(e) {
      $box.append("<p class='bold'>Connection closed!</p>");
    };
    ws.onmessage = function(e) {
      $box.append("<p>"+e.data+"</p>");
      increaseUnreadCount();
    };

    ws.onerror = function(e) {
      $box.append("<strong>Error!</strong>")
    };

    $("#form").submit(function(e) {
      e.preventDefault();
      if (!ws) {
          return;
      }
      if (!$msg.val()) {
          return;
      }
      ws.send($msg.val());
      $msg.val("");
    });

    document.addEventListener("visibilitychange", resetUnreadCount);

    function increaseUnreadCount() {
      if (document.hidden === true) {
        var count = parseInt(document.title.match(/\d+/));
        if (!count) {
          document.title = "(1) Chat";
          return;
        }
        document.title = "("+(count+1)+") Chat";
      }
    }

    function resetUnreadCount() {
      if (document.hidden === false) {
        document.title = "Chat";
      }
    }

  });
  </script>
  </body>
</html>
`

var (
	maxMsgLen          = 10 // @TODO: IMPLEMENT THESE. server should validate
	maxNameLen         = 40
	errMessageTooLong  = errors.New("Messages must be less than 10 characters")
	errUsernameTooLong = errors.New("Usernames cannot be more than 40 charcters")
)

type server struct {
	seq        uint64
	logger     *log.Logger
	clients    map[uint64]client
	newConn    chan net.Conn
	newWsConn  chan *websocket.Conn
	msgRcv     chan string
	disconnect chan client
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(homeHTML))
}

func (s *server) serve(port string) error {
	server, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	s.logger.Println("Server started on ", port)

	// TCP Server
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				s.logger.Println(err.Error())
			}
			s.newConn <- conn
		}
	}()

	// HTTP Server/Websocket server
	go func() {
		http.HandleFunc("/", homeHandler)
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			println("DID SOMETHING")
			newWsClientHandler(s, w, r)
		})
		http.ListenAndServe(":8000", nil)
	}()

	for {
		select {
		// this should probably be it's own function soon it's pretty involved
		case conn := <-s.newConn:
			s.seq++
			c := &tcpClient{
				id:     s.seq,
				name:   strconv.Itoa(int(s.seq)),
				r:      bufio.NewReader(conn),
				w:      bufio.NewWriter(conn),
				conn:   conn,
				server: s,
			}
			s.clients[c.id] = c
			c.write(chatHelp)
			s.broadcast("(chatbot): New user joined\n")
			go c.read()

		case conn := <-s.newWsConn:
			s.seq++
			ws := &wsClient{
				id:     s.seq,
				name:   strconv.Itoa(int(s.seq)),
				conn:   conn,
				server: s,
			}
			s.clients[ws.id] = ws
			ws.write(chatHelp)
			s.broadcast("(chatbot): New user joined\n")
			go ws.read()

		case msg := <-s.msgRcv:
			s.logger.Print("Message received: ", msg)
			s.broadcast(msg)

		case c := <-s.disconnect:
			s.logger.Printf("Disconnected user %s\n", c.getName())
			s.broadcast(fmt.Sprintf("(chatbot): user %s left the chat\n", c.getName()))
			delete(s.clients, c.getID())
			c.close()
		}
	}
}

// broadcast is the function to use to handle broadcasting to multiple
// rooms n stuff
func (s *server) broadcast(msg string) {
	for _, c := range s.clients {
		err := c.write(msg)
		if err != nil {
			s.logger.Println("Broadcast error: ", err.Error())
		}
	}
}

func ServeTCP(l *log.Logger, port string) error {
	s := &server{
		logger:     l,
		clients:    make(map[uint64]client),
		newConn:    make(chan net.Conn),
		newWsConn:  make(chan *websocket.Conn),
		msgRcv:     make(chan string),
		disconnect: make(chan client),
	}
	return s.serve(port)
}
