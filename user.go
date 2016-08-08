package torbit

import (
	"net"
	"net/http"
)

type connection interface {
	read() error
	write(message *message) error
	close()
}

// A User represents a user in the chat. Their connection is used to
// communicate
type User struct {
	name string
	conn connection
}

func createTCPUser(conn net.Conn, h *hub, w http.ResponseWriter) *User {
	u := newTCPUser(conn, h)
	u.write(newMessage(u.currentRoomName, u.username, chatHelp, text))
	return &User{
		name: u.name(),
		conn: u,
	}
}

func createWSUser(w http.ResponseWriter, r *http.Request, h *hub) *User {
	u, err := newWsUser(w, r, h)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	u.write(newMessage(u.currentRoomName, u.username, chatHelp, text))
	return &User{
		name: u.username,
		conn: u,
	}
}
