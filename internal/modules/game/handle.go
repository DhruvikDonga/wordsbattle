package game

import (
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
	Wordslist        map[string]bool
	ClientProperties map[string]*ClientProps
	Rounds           int
}

type GameRoomData struct {
	IsRandomGame     bool
	PlayerLimit      int
	ClientScore      map[string]int
	Wordslist        map[string]bool
	ClientProperties map[string]*ClientProps
	Rounds           int
}

func (r *RoomData) HandleRoomData(room gogamemesh.Room, server gogamemesh.MeshServer) {
	roomname := room.GetRoomSlugInfo()
	log.Println("Handeling data for ", roomname, server.GetGameName())
	// ticker := time.NewTicker(1 * time.Second)
	// defer ticker.Stop()
	for {
		select {
		case message, ok := <-room.ConsumeRoomMessage():
			//log.Println(server.GetRooms())
			if !ok {
				log.Println("Channel closed. Exiting HandleRoomData for", roomname)
			}
			log.Println("Room data ", message)

			if message.Target == roomname {
				r.handleServermessages(room, server, message)
			}

		case clientevent := <-room.EventTriggers():
			log.Println("Event triggered", clientevent[0], clientevent[1], clientevent[2], room.GetRoomSlugInfo())

		// case <-ticker.C:
		// 	log.Println("Room activity", roomname)

		case <-room.RoomStopped():
			log.Println("Room is stopped so stop the handler")
			return

		}
	}
}

// global room
func (r *RoomData) handleServermessages(room gogamemesh.Room, server gogamemesh.MeshServer, message *gogamemesh.Message) {

	// Unmarshal the JSON data into the map
	log.Println("game name:-", server.GetGameName(), message.Action)
	switch message.Action {
	case "join-room":
		rd := &GameRoomData{
			IsRandomGame: false,
			PlayerLimit:  int(message.MessageBody["playerlimit"].(float64)),
		}
		log.Println("JoinRoomAction ", message.Sender, message.MessageBody, room.GetRoomSlugInfo())
		roomname := message.MessageBody["roomname"].(string)
		//var joinroomrequest []interface{}
		//joinroomrequest = append(joinroomrequest, roomname, message.Sender, rd)
		//server.JoinARoom() <- joinroomrequest
		server.JoinClientRoom(roomname, message.Sender, rd)
		log.Println("request send to join a room")
		// message := &gogamemesh.Message{
		// 	Action: "join-room-notify",
		// 	Target: message.MessageBody["roomname"].(string),
		// 	MessageBody: map[string]interface{}{
		// 		"newmessage": fmt.Sprintf("%s %s joined the room cowgame by mesh", r.ClientProperties[message.Sender].Name, message.Sender),
		// 	},
		// 	Sender:         message.Sender,
		// 	IsTargetClient: false,
		// }

		// server.BroadcastMessage(message)

	}
}

// game room
func (r *GameRoomData) handleGameRoommessages(room gogamemesh.Room, server gogamemesh.MeshServer, message *gogamemesh.Message) *gogamemesh.Message {
	ret := &gogamemesh.Message{}
	return ret

}

func (r *GameRoomData) HandleRoomData(room gogamemesh.Room, server gogamemesh.MeshServer) {
	roomname := room.GetRoomSlugInfo()
	log.Println("Handeling data for ", roomname)
	for {
		select {
		case message := <-room.ConsumeRoomMessage():

			if message.Target == roomname {
				ret := r.handleGameRoommessages(room, server, message)
				server.BroadcastMessage(ret)
			}

		case clientsinroom := <-room.EventTriggers():
			log.Println("Event triggered", clientsinroom[0], clientsinroom[1], clientsinroom[2])
			switch clientsinroom[0] {
			case "client-joined-room":
				message := &gogamemesh.Message{
					Action: "join-room-notify",
					Target: clientsinroom[1],
					MessageBody: map[string]interface{}{
						"newmessage": fmt.Sprintf("%s %s joined the room cowgame by mesh", r.ClientProperties[clientsinroom[2]].Name, clientsinroom[2]),
					},
					Sender:         clientsinroom[2],
					IsTargetClient: false,
				}

				server.BroadcastMessage(message)
			}
		case <-room.RoomStopped():
			log.Println("Room is stopped so stop the handler")
			return
		}
	}
}
