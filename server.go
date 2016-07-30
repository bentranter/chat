package torbit

import (
	"bufio"
	"log"
	"net"
)

func ServeTCP(l *log.Logger, port string) error {
	s := &server{
		logger:      l,
		clientCount: 0,
		clients:     make(map[net.Conn]int),
		newConn:     make(chan net.Conn),
		msgRcv:      make(chan string),
	}
	return s.serve(port)
}

type server struct {
	logger      *log.Logger
	clientCount int              // should be unique ID
	clients     map[net.Conn]int // also unique ID
	newConn     chan net.Conn    // connections received here
	msgRcv      chan string      // msgs are received here
}

// listen is the TCP server that listens
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
			s.clientCount++
			s.clients[conn] = s.clientCount
			go s.handleConn(conn, s.clientCount)

		case msg := <-s.msgRcv:
			s.logger.Print("Message reveived: ", msg)
			for conn := range s.clients {
				go s.broadcast(conn, msg) // scared of leaking goroutines...
			}
		}
	}
}

func (s *server) handleConn(conn net.Conn, clientID int) {
	// read each message and jam it in the messages channel
	r := bufio.NewReader(conn)
	for {
		in, err := r.ReadString('\n')
		if err != nil {
			// should signal that someone disconnected
			s.logger.Println(err.Error())
			break
		}
		// use the clientID, but they should have a name
		s.msgRcv <- string(clientID) + in
	}

	// if we've errored, we'll be out of the loop for that conn, so close the
	// connection
	conn.Close()
}

func (s *server) broadcast(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		s.logger.Println(err.Error())
		// send something to disconnect channel or else you'll leak goroutines
	}
}
