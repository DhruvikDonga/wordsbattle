package handler

import (
	"net/http"

	rs "github.com/DhruvikDonga/wordsbattle/pkg"
)

type HelloMessage struct {
	Description string `json:"description"`
}

func (app *App) HelloHandler(w http.ResponseWriter, r *http.Request) {
	type messages []HelloMessage
	res := messages{
		HelloMessage{Description: "WordsBattle❤"},
		HelloMessage{Description: "Welcome to coolest project ever"},
	}

	rs.RespondwithJSON(w, http.StatusOK, res)
}
