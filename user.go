package torbit

import (
	"net"
	"net/http"

	"github.com/julienschmidt/httprouter"
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

func createTCPUser(conn net.Conn, h *hub) *User {
	u := newTCPUser(conn, h)
	u.write(newMessage(u.currentRoomName, u.username, chatHelp, text))
	return &User{
		name: u.name(),
		conn: u,
	}
}

func createWSUser(h *hub, w http.ResponseWriter, r *http.Request, _ httprouter.Params) *User {
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
