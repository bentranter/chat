package torbit

import (
	"bufio"
	"net"
	"strings"
)

type tcpClient struct {
	name   string
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
		if ok := s.clients[n]; ok == nil {
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
		r:      r,
		w:      w,
		conn:   conn,
		server: s,
	}
}

func (c *tcpClient) getName() string {
	return c.name
}

func (c *tcpClient) read() {
	for {
		msg, err := c.r.ReadString('\n')
		if err != nil {
			c.server.disconnect <- c
			break
		}
		if ok := handleCommand(c, msg); ok {
			continue
		}
		c.server.msgRcv <- "(" + c.name + "): " + msg
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
