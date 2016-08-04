package torbit

// import (
// 	"net/http"

// 	"github.com/gorilla/websocket"
// )

// type wsClient struct {
// 	name   string
// 	room   string
// 	conn   *websocket.Conn
// 	server *server
// }

// func homeHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "text/html")
// 	w.Write([]byte(homeHTML))
// }

// func newWsClientHandler(s *server, w http.ResponseWriter, r *http.Request) {
// 	wsconn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	var name string
// 	wsconn.WriteMessage(websocket.TextMessage, []byte("(chatbot): Please enter your username.\n "))
// 	for {
// 		_, n, err := wsconn.ReadMessage()
// 		if err != nil {
// 			s.logger.Println("Error reading from new websocket connection: ", err.Error())
// 			wsconn.Close()
// 		}
// 		username := string(n)
// 		if _, ok := s.clients[username]; !ok {
// 			name = string(username)
// 			break
// 		}
// 		wsconn.WriteMessage(websocket.TextMessage, []byte("(chatbot): Sorry, the name "+username+" is already taken. Please choose another one.\n"))
// 	}

// 	ws := &wsClient{
// 		name:   name,
// 		room:   defaultRoomName,
// 		conn:   wsconn,
// 		server: s,
// 	}
// 	s.join <- ws
// }

// func (ws *wsClient) getName() string {
// 	return ws.name
// }

// func (ws *wsClient) getRoom() string {
// 	return ws.room
// }

// func (ws *wsClient) setRoom(room string) {
// 	ws.room = room
// }

// func (ws *wsClient) read() {
// 	for {
// 		_, msg, err := ws.conn.ReadMessage()
// 		if err != nil {
// 			ws.server.leave <- ws
// 			break
// 		}
// 		if ok := handleCommand(ws.server, ws, string(msg)); ok {
// 			continue
// 		}
// 		ws.server.recv <- &message{
// 			content:  "(" + ws.name + "): " + string(msg) + "\n",
// 			roomName: ws.room,
// 		}
// 	}
// }

// func (ws *wsClient) write(msg string) error {
// 	return ws.conn.WriteMessage(websocket.TextMessage, []byte(msg))
// }

// func (ws *wsClient) close() {
// 	ws.conn.Close()
// }
