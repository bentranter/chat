package torbit

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func getServeMux(h *hub) http.Handler {
	r := httprouter.New()

	r.GET("/", homeHandler)
	r.POST("/messages", handle(h, newMessageHandler))
	r.POST("/ws", handle(h, createWSUserHandler))

	return r
}

type handler func(h *hub, w http.ResponseWriter, r *http.Request, ps httprouter.Params)

func handle(hub *hub, h handler) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		h(hub, w, r, ps)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Nice"))
}

func createWSUserHandler(h *hub, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	u := createWSUser(h, w, r, nil)
	h.userCh <- u
}

func newMessageHandler(h *hub, w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	msg := &message{}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	h.messageCh <- msg
}
