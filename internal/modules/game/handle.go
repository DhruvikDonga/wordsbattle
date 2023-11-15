package game

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/DhruvikDonga/wordsbattle/pkg/gogamemesh"
)

const (
	JoinRoomAction    = "join-room"
	LeaveRoomAction   = "leave-room"
	PushMessageAction = "push-message"
)

const (
	ClientJoinedNotification       = "client-joined-room"
	ClientLeftNotification         = "client-left-room"
	ClientConnectedNotification    = "client-connected"
	ClientDisconnectedNotification = "client-disconnected"
)

type ClientProps struct { //this can depend on a inroom basis so it changes
	Color string
	Name  string
}

type RoomData struct {
	IsRandomGame     bool
	PlayerLimit      int
	ClientScore      map[string]int
	Wordslist        map[string]bool
	ClientProperties map[string]*ClientProps
	Rounds           int
}

func (r *RoomData) HandleRoomData(room gogamemesh.Room, server gogamemesh.MeshServer) {
	roomname := room.GetRoomSlugInfo()
	select {
	case message := <-server.RecieveMessage():
		if message.Target == roomname && roomname == gogamemesh.MeshGlobalRoom {
			log.Println(message)
			ret := r.handleServermessages(room, server, message)
			server.BroadcastMessage(ret)
		}

	case clientsinroom := <-server.EventTriggers():
		log.Println(clientsinroom[roomname])

	case <-room.RoomStopped():
		log.Println("Room is stopped so stop the handler")
		return
	}
}

func (r *RoomData) handleServermessages(room gogamemesh.Room, server gogamemesh.MeshServer, message *gogamemesh.Message) *gogamemesh.Message {
	ret := &gogamemesh.Message{}
	var messagebody map[string]interface{}

	// Unmarshal the JSON data into the map
	err := json.Unmarshal(message.MessageBody, &messagebody)
	if err != nil {
		fmt.Println("Error:", err)
	}
	switch message.Action {
	case JoinRoomAction:
		rd := &RoomData{
			IsRandomGame: false,
			PlayerLimit:  messagebody["playerlimit"].(int),
		}
		server.JoinClientRoom(messagebody["roomname"].(string), message.Sender, rd)

	}
	return ret
}
