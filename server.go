package torbit

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const chatHelp = `(chatbot): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /name    set your name               (example: /name Ben)
  /id      view your user id           (example: /id)

`

var (
	maxMsgLen          = 10 // @TODO: IMPLEMENT THESE
	maxNameLen         = 40
	errMessageTooLong  = errors.New("Messages must be less than 10 characters")
	errUsernameTooLong = errors.New("Usernames cannot be more than 40 charcters")
)

type client struct {
	id     uint64
	name   string
	r      *bufio.Reader
	w      *bufio.Writer
	conn   net.Conn
	server *server
}

func (c *client) read() {
	for {
		msg, err := c.r.ReadString('\n')
		if err != nil {
			c.server.disconnect <- c
			break
		}
		// please make this a function (or func map) yikes
		if strings.HasPrefix(msg, "/help") {
			c.write(chatHelp)
			continue
		}
		if strings.HasPrefix(msg, "/name") {
			name := strings.TrimSpace(strings.TrimLeft(msg, "/name "))
			if name == "" {
				c.write("(chatbot): Your name can't be empty\n")
				continue
			}
			c.name = name
			c.write("(chatbot): Your name is " + name + "\n")
			continue
		}
		if strings.HasPrefix(msg, "/id") {
			idMsg := fmt.Sprintf("(chatbot): Your id is %d\n", c.id)
			c.write(idMsg)
			continue
		}

		c.server.msgRcv <- "(" + c.name + "): " + msg
	}
}

func (c *client) write(msg string) error {
	_, err := c.w.WriteString(msg)
	if err != nil {
		return err
	}
	return c.w.Flush()
}

// In order for the server to be cool with TCP and WS based clients,
// the client here might need to be an interface
type server struct {
	seq        uint64
	logger     *log.Logger
	clients    map[uint64]*client
	newConn    chan net.Conn
	msgRcv     chan string
	disconnect chan *client // disconnect via this channel, could be int after change above ismade
}

// listen is the TCP server that listens
//
// it's less coupled now but the interface idea above might be the best to reduce
// coupling. its also a bit too complicated
func (s *server) serve(port string) error {
	server, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	s.logger.Println("Server started on ", port)

	// accept new conns
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				s.logger.Println(err.Error())
			}
			s.newConn <- conn
		}
	}()

	for {
		select {
		// this should probably be it's own function soon it's pretty involved
		case conn := <-s.newConn:
			s.seq++
			c := &client{
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

		case msg := <-s.msgRcv:
			s.logger.Print("Message received: ", msg)
			s.broadcast(msg)

		// delete disconnected clients
		case c := <-s.disconnect:
			s.logger.Printf("Disconnected user %s\n", c.name)
			s.broadcast(fmt.Sprintf("(chatbot): user %s left the chat\n", c.name))
			delete(s.clients, c.id) // remove user
			c.conn.Close()
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
		clients:    make(map[uint64]*client),
		newConn:    make(chan net.Conn),
		msgRcv:     make(chan string),
		disconnect: make(chan *client),
	}
	return s.serve(port)
}
