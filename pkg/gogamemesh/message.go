package gogamemesh

import (
	"encoding/json"
	"log"
)

// Message struct is the structure of the message which is send in mesh server
type Message struct {
	Action         string `json:"action"`       //action
	MessageBody    []byte `json:"message_body"` //message
	IsTargetClient bool   //not imported if its true then the Target string is a client which is one
	Target         string `json:"target"` //target the room
	Sender         string `json:"sender"` //whose readpump is used
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
