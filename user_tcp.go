package torbit

import (
	"bufio"
	"net"
	"strings"
	"time"
)

const chatHelp = `(chatbot): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /rooms   list all rooms              (example: /rooms)
  /join    join a room                 (example: /join general)
  /newroom create a new room           (example: /newroom random)

`

type tcpUser struct {
	currentRoomName string
	username        string
	r               *bufio.Reader
	w               *bufio.Writer
	conn            net.Conn
	receiver        chan *message
}

func newTCPUser(conn net.Conn, receiver chan *message) *tcpUser {
	// in here, read name, etc
	return &tcpUser{
		currentRoomName: defaultRoomName,
		r:               nil,
		w:               nil,
		conn:            conn,
		receiver:        receiver,
	}
}

func (tc *tcpUser) read() error {
	for {
		messageText, err := tc.r.ReadString('\n')
		if err != nil {
			return err
		}
		// handle commands here to determine:
		//   1. text (duh)
		//   2. messageType
		tc.receiver <- &message{
			channel:     tc.currentRoomName,
			username:    tc.username,
			text:        messageText,
			time:        time.Now(),
			messageType: text,
		}
	}
}

func (tc *tcpUser) write(message string) error {
	_, err := tc.w.WriteString(message)
	if err != nil {
		return err
	}
	return tc.w.Flush()
}

func (tc *tcpUser) close() {
	tc.conn.Close()
}

func (tc *tcpUser) name() string {
	return tc.username
}

func (tc *tcpUser) setCurrentRoom(roomName string) {
	tc.currentRoomName = roomName
}

type tcpClient struct {
	name   string
	room   string
	r      *bufio.Reader
	w      *bufio.Writer
	conn   net.Conn
	server *server
}

func newTCPClient(conn net.Conn, s *server) *tcpClient {
	var name string
	w := bufio.NewWriter(conn)
	r := bufio.NewReader(conn)
	w.WriteString("(chatbot): Please enter your username: ")
	w.Flush()

	for {
		n, err := r.ReadString('\n')
		if err != nil {
			s.logger.Println("Error reading from new client: ", err.Error())
			conn.Close()
		}
		n = strings.TrimSpace(n)
		if _, ok := s.clients[n]; !ok { // they already online baby
			name = n
			break
		}
		// name is already taken
		w.WriteString("(chatbot): Sorry, the name " + n + " is already taken. Please choose another one: ")
		w.Flush()
		continue
	}

	return &tcpClient{
		name:   name,
		room:   defaultRoomName,
		r:      r,
		w:      w,
		conn:   conn,
		server: s,
	}
}

func (c *tcpClient) getName() string {
	return c.name
}

func (c *tcpClient) getRoom() string {
	return c.room
}

func (c *tcpClient) setRoom(room string) {
	c.room = room
}

func (c *tcpClient) read() {
	for {
		msg, err := c.r.ReadString('\n')
		if err != nil {
			c.server.leave <- c
			break
		}
		if ok := handleCommand(c.server, c, msg); ok {
			continue
		}
		c.server.recv <- &message{
			content:  "(" + c.name + "): " + msg,
			roomName: c.room,
		}
	}
}

func (c *tcpClient) write(msg string) error {
	_, err := c.w.WriteString(msg)
	if err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *tcpClient) close() {
	c.conn.Close()
}
