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

func createTCPUser(conn net.Conn, receiver chan *message) *User {
	u := newTCPUser(conn, receiver)
	return &User{
		name: u.name(),
		conn: u,
	}
}

// func createWSUser() {}
