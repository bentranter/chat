package torbit

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func getServeMux() http.Handler {
	r := httprouter.New()

	r.GET("/", homeHandler)
	r.POST("/messages", newMessageHandler)
	// r.POST("/ws", wsHandler)

	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Write([]byte("Nice"))
}

func newMessageHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	s := &struct {
		Test string
	}{}
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	w.Write([]byte(s.Test))
}
