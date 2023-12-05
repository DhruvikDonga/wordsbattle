package game

import (
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

type RoomData struct {
	Slug             string
	Wordslist        map[string]bool
	ClientProperties map[string]*ClientProps
	Rounds           int
}

func (r *RoomData) HandleRoomData(room gogamemesh.Room, server gogamemesh.MeshServer) {
	roomname := room.GetRoomSlugInfo()
	r.Slug = roomname
	log.Println("Handeling data Server for ", roomname, server.GetGameName())
	// ticker := time.NewTicker(5 * time.Second)
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
		// 	log.Println("Room activity server", room.GetRoomSlugInfo(), r.Slug)

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
		//needed only if new room is needed
		rd := GameRoomData{
			IsRandomGame:     false,
			PlayerLimit:      int(message.MessageBody["playerlimit"].(float64)),
			ClientProperties: make(map[string]*ClientProps),
		}
		log.Println("JoinRoomAction ", message.Sender, message.MessageBody, room.GetRoomSlugInfo())
		roomname := message.MessageBody["roomname"].(string)
		//var joinroomrequest []interface{}
		//joinroomrequest = append(joinroomrequest, roomname, message.Sender, rd)
		//server.JoinARoom() <- joinroomrequest
		server.JoinClientRoom(roomname, message.Sender, &rd)
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
