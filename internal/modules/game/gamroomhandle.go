package game

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/DhruvikDonga/wordsbattle/pkg/gogamemesh"
)

type GameRoomData struct {
	mu               sync.RWMutex
	Slug             string
	IsRandomGame     bool
	PlayerLimit      int
	ClientScore      map[string]int
	Wordslist        map[string]bool
	ClientProperties map[string]*ClientProps
	Rounds           int
}
type ClientProps struct { //this can depend on a inroom basis so it changes
	Color string `json:"color"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Slug  string `json:"slug"`
}

func (r *GameRoomData) HandleRoomData(room gogamemesh.Room, server gogamemesh.MeshServer) {
	roomname := room.GetRoomSlugInfo()
	r.Slug = room.GetRoomSlugInfo()

	log.Println("Handeling data for ", roomname)
	//ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case message := <-room.ConsumeRoomMessage():

			if message.Target == roomname {
				r.handleGameRoommessages(room, server, message)
			}

		case clientsinroom := <-room.EventTriggers():
			log.Println("Event triggered", clientsinroom[0], clientsinroom[1], clientsinroom[2])
			switch clientsinroom[0] {
			case "client-joined-room":
				r.JoinRoomNotify(clientsinroom, room, server) //to all in the room
				r.KnowTheClient(clientsinroom, room, server)  //to only the client
			case "client-left-room":
				r.ClientListNotify(clientsinroom, room, server)

			}
		// case <-ticker.C:
		// 	log.Println("Room activity game room", room.GetRoomSlugInfo(), r.Slug)
		case <-room.RoomStopped():
			log.Println("Room is stopped so stop the handler")
			return
		}
	}
}

// game room
func (r *GameRoomData) handleGameRoommessages(room gogamemesh.Room, server gogamemesh.MeshServer, message *gogamemesh.Message) {
	switch message.Action {
	case "set-client-name":
		colors := []string{"yellow", "red", "orange", "blue", "purple", "pink", "white"}
		color := colors[rand.Intn(len(colors))]
		r.ClientProperties[message.Sender] = &ClientProps{Name: message.MessageBody["setname"].(string), Color: color, Score: 0}
		log.Printf("Set client props: %+v\n ", r.ClientProperties[message.Sender])
		clientsinroom := []string{"client-list-notify", message.Target, message.Sender}
		r.ClientListNotify(clientsinroom, room, server)
	}

}

type ClientsinRoomMessage struct { // we are using this to return list of clients to all clients in room when register unregister happens
	ClientList []*ClientProps `json:"clientlist"` //message
}

func (r *GameRoomData) ClientListNotify(clientsinroom []string, room gogamemesh.Room, server gogamemesh.MeshServer) {
	ret := []*ClientProps{}
	if clientsinroom[0] == "client-left-room" {
		r.mu.Lock()
		delete(r.ClientProperties, clientsinroom[2])
		r.mu.Unlock()
	}
	for slug, props := range r.ClientProperties {
		temp := &ClientProps{
			Color: props.Color,
			Name:  props.Name,
			Score: props.Score,
			Slug:  slug,
		}
		ret = append(ret, temp)
	}
	message := &gogamemesh.Message{
		Action: "client-list-notify",
		Target: clientsinroom[1],
		MessageBody: map[string]interface{}{
			"clientsinroomessage": ret,
		},
		Sender:         "Gawd",
		IsTargetClient: false,
	}

	server.BroadcastMessage(message)
}

func (r *GameRoomData) KnowTheClient(clientsinroom []string, room gogamemesh.Room, server gogamemesh.MeshServer) {
	message := &gogamemesh.Message{
		Action:         "know-yourself",
		Target:         clientsinroom[1],
		Sender:         clientsinroom[2],
		IsTargetClient: true,
	}

	server.BroadcastMessage(message)
}

func (r *GameRoomData) JoinRoomNotify(clientsinroom []string, room gogamemesh.Room, server gogamemesh.MeshServer) {
	message := &gogamemesh.Message{
		Action: "join-room-notify",
		Target: clientsinroom[1],
		MessageBody: map[string]interface{}{
			"newmessage": fmt.Sprintf("%s joined the room cowgame by mesh", clientsinroom[2]),
		},
		Sender:         clientsinroom[2],
		IsTargetClient: false,
	}

	server.BroadcastMessage(message)
}
