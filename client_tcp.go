package torbit

import (
	"bufio"
	"net"
	"strings"
)

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

func (c *tcpClient) roomChangeCh() chan *roomChange {
	return c.server.change // dumb
}

func (c *tcpClient) read() {
	for {
		msg, err := c.r.ReadString('\n')
		if err != nil {
			c.server.leave <- c
			break
		}
		if ok := handleCommand(c, msg); ok {
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
