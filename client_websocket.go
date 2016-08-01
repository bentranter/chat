package torbit

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type wsClient struct {
	id     uint64
	name   string
	conn   *websocket.Conn
	server *server
}

func newWsClientHandler(s *server, w http.ResponseWriter, r *http.Request) {
	wsconn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.newWsConn <- wsconn
}

func (ws *wsClient) getID() uint64 {
	return ws.id
}

func (ws *wsClient) getName() string {
	return ws.name
}

func (ws *wsClient) setName(name string) {
	ws.name = name
}

func (ws *wsClient) read() {
	for {
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			ws.server.disconnect <- ws
			break
		}
		if ok := handleCommand(ws, string(msg)); ok {
			continue
		}
		ws.server.msgRcv <- "(" + ws.name + "): " + string(msg) + "\n"
	}

}

func (ws *wsClient) write(msg string) error {
	return ws.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (ws *wsClient) close() {
	ws.conn.Close()
}
