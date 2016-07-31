package torbit

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func ServeTCP(l *log.Logger, port string) error {
	s := &server{
		logger:     l,
		clients:    []*client{},
		newConn:    make(chan net.Conn),
		msgRcv:     make(chan string),
		disconnect: make(chan *client),
	}
	return s.serve(port)
}

// i don't like how tighltly coupled this is..
type client struct {
	name   string
	conn   net.Conn
	server *server
}

func (c *client) read() {
	// read each message and jam it in the messages channel
	r := bufio.NewReader(c.conn)
	for {
		in, err := r.ReadString('\n')
		if err != nil {
			c.server.disconnect <- c
			break
		}
		if strings.HasPrefix(in, "/help") {
			c.conn.Write([]byte("Usage\n\t/name\tSets you set your name (example: /name Ben)\n\t/help\tShows this help message again\n\n"))
			continue
		}
		if strings.HasPrefix(in, "/name") {
			name := strings.TrimSpace(strings.TrimLeft(in, "/name "))
			if name == "" {
				c.conn.Write([]byte("Your name can't be empty\n"))
				continue
			}
			c.name = name
			c.conn.Write([]byte("Your name is " + name + "\n"))
			continue
		}
		c.server.msgRcv <- "(" + c.name + "): " + in
	}
}

type server struct {
	logger     *log.Logger
	clients    []*client     // yay lcients get to maintain their own connection
	newConn    chan net.Conn // connections received here
	msgRcv     chan string   // msgs are received here
	disconnect chan *client  // disconnect via this channel
}

// listen is the TCP server that listens
//
// this is really tightly coupled with the client. it'd be nice if the client
// and the server didn't have to care about each other.
//
// maybe the disconnect channel can be moved from the server to the client...
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
		case conn := <-s.newConn:
			c := &client{
				name:   "anonymous",
				conn:   conn,
				server: s,
			}
			s.clients = append(s.clients, c)
			// Write a hello message to their connection
			conn.Write([]byte("\nHello! Welcome to the chat room!\nType /help to see available commands, or type and message to chat!\n\n"))
			go c.read()

		case msg := <-s.msgRcv:
			s.logger.Print("Message received: ", msg)
			for _, client := range s.clients {
				s.broadcast(client.conn, msg) // might not need a goroutine?
			}

		// delete disconnected clients
		case c := <-s.disconnect:
			s.logger.Printf("Disconnected user %s\n", c.name)
			// need to remove user
			c.conn.Close()
		}
	}
}

func (s *server) broadcast(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		s.logger.Println("Broadcast error: ", err.Error())
	}
}
