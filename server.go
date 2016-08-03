package torbit

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

const (
	defaultRoomName = "general"
	chatHelp        = `(chatbot): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /join    join a room (example: /join general)

`
)

var (
	errAlreadyInRoom    = errors.New("You're already in that room")
	errRoomDoesNotExist = errors.New("Room doesn't exist")
	errRoomExists       = errors.New("Room already exists")
)

type message struct {
	content  string
	roomName string
}

// needs to be a struct with a mutex since it's accessed by goroutines
type room map[string]client // client name -> client

type roomChange struct {
	newRoomName string
	c           client
}

// needs mutex
type server struct {
	logger  *log.Logger
	clients map[string]bool // only for keeping track usernames
	rooms   map[string]room // room name -> room
	join    chan client
	recv    chan *message
	change  chan *roomChange
	leave   chan client
}

// @TODO: something needs to be done about this horrible mess
func (s *server) newClient(c client) {
	s.clients[c.getName()] = true             // client joins server, name is reserved
	s.rooms[defaultRoomName][c.getName()] = c // jesus christ lol, client joins room
	c.write(chatHelp)
	welcomeMessage := &message{
		content:  "(chatbot): New user " + c.getName() + " has joined.\n",
		roomName: defaultRoomName,
	}
	s.broadcast(welcomeMessage)
	go c.read()
}

// newRoom creates a new room with the given name, and adds that room to the
// list of room maintained by the server.
func (s *server) newRoom(name string) error {
	if _, ok := s.rooms[name]; ok {
		return errRoomExists
	}
	s.rooms[name] = make(room)
	return nil
}

func (s *server) changeRoom(r *roomChange) error {
	fmt.Printf("Rooms: %#v\n", s.rooms)
	if _, ok := s.rooms[strings.TrimSpace(r.newRoomName)]; !ok {
		return errRoomDoesNotExist
	}

	// join the other room
	if _, ok := s.rooms[r.newRoomName][r.c.getName()]; ok {
		return errAlreadyInRoom
	}

	delete(s.rooms[r.c.getRoom()], r.c.getName()) // remove from current room
	s.rooms[r.newRoomName][r.c.getName()] = r.c
	r.c.setRoom(r.newRoomName)
	return nil
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
			s.join <- newTCPClient(conn, s)
		}
	}()

	// HTTP Server/Websocket server
	go func() {
		http.HandleFunc("/", homeHandler)
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			newWsClientHandler(s, w, r)
		})
		http.ListenAndServe(":8000", nil)
	}()

	for {
		select {
		case client := <-s.join:
			s.newClient(client)

		case msg := <-s.recv:
			s.logger.Print("Message received: ", msg.content)
			s.broadcast(msg)

		case changeReq := <-s.change:
			err := s.changeRoom(changeReq)
			if err != nil {
				changeReq.c.write(err.Error() + "\n")
				continue
			}
			changeReq.c.write("Changed to room " + changeReq.newRoomName + "\n")

		case c := <-s.leave:
			s.logger.Printf("Disconnected user %s\n", c.getName())
			s.broadcast(&message{
				content:  fmt.Sprintf("(chatbot): user %s left the chat\n", c.getName()),
				roomName: c.getRoom(),
			})
			delete(s.clients, c.getName())
			delete(s.rooms[c.getRoom()], c.getName()) // jesus christ
			c.close()
		}
	}
}

// broadcast writes a message to the given room
func (s *server) broadcast(m *message) {
	room, ok := s.rooms[m.roomName]
	if !ok {
		return
		// room doesn't exist so idk exactly what to do - maybe err?
	}
	for _, c := range room {
		err := c.write(m.content)
		if err != nil {
			s.logger.Println("Broadcast error: ", err.Error())
		}
	}
}

func ServeTCP(l *log.Logger, port string) error {
	s := &server{
		logger:  l,
		clients: make(map[string]bool),
		rooms:   make(map[string]room),
		join:    make(chan client),
		recv:    make(chan *message),
		change:  make(chan *roomChange),
		leave:   make(chan client),
	}
	err := s.newRoom(defaultRoomName)
	if err != nil {
		return err
	}
	err = s.newRoom("chat")
	if err != nil {
		return err
	}
	return s.serve(port)
}

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
