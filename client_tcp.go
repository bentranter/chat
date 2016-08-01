package torbit

import (
	"bufio"
	"net"
)

type tcpClient struct {
	id     uint64
	name   string
	r      *bufio.Reader
	w      *bufio.Writer
	conn   net.Conn
	server *server
}

func (c *tcpClient) getID() uint64 {
	return c.id
}

func (c *tcpClient) getName() string {
	return c.name
}

func (c *tcpClient) setName(name string) {
	c.name = name
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
