package game

import (
	"log"

	"github.com/DhruvikDonga/wordsbattle/pkg/gogamemesh"
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
		if message.Target == roomname {
			log.Println(message)
		}
		server.PushMessage() <- message

	case clientsinroom := <-server.EventTriggers():
		log.Println(clientsinroom[roomname])

	case <-room.RoomStopped():
		log.Println("Room is stopped so stop the handler")
		return
	}
}
