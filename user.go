package torbit

import (
	"net"
)

type connection interface {
	read() error
	write(message string) error
	close()
}

// A User represents a user in the chat. Their connection is used to
// communicate
type User struct {
	name string
	conn connection
}

func createTCPUser(conn net.Conn, h *hub) *User {
	u := newTCPUser(conn, h)
	u.write(chatHelp)
	return &User{
		name: u.name(),
		conn: u,
	}
}

// func createWSUser() {}
