package torbit

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

var errNameNotAvailable = errors.New("That name is not available")

type wsUser struct {
	currentRoomName string
	muted           map[string]bool
	username        string
	conn            *websocket.Conn
	send            chan<- *message
}

func newWsUser(w http.ResponseWriter, r *http.Request, h *hub) (*wsUser, error) {
	user := &struct {
		Name string
	}{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if _, ok := h.users[user.Name]; ok {
		return nil, errNameNotAvailable
	}

	wsconn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		return nil, err
	}

	return &wsUser{
		currentRoomName: defaultChannelName,
		muted:           make(map[string]bool),
		username:        user.Name,
		conn:            wsconn,
		send:            h.messageCh,
	}, nil
}

func (ws *wsUser) read() error {
	for {
		msg := &message{}
		err := ws.conn.ReadJSON(msg)
		if err != nil {
			// not sure how to handle this based on severity...
			continue
		}
		ws.send <- msg
	}
}

func (ws *wsUser) write(message *message) error {
	return ws.conn.WriteJSON(message)
}

func (ws *wsUser) close() {
	ws.conn.Close()
}
