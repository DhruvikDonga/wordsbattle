package game

import (
	"log"

	"github.com/DhruvikDonga/simplysocket"
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
	RandomRooms      []string
}

func (r *RoomData) HandleRoomData(room simplysocket.Room, server simplysocket.MeshServer) {
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
func (r *RoomData) handleServermessages(room simplysocket.Room, server simplysocket.MeshServer, message *simplysocket.Message) {

	// Unmarshal the JSON data into the map
	log.Println("game name:-", server.GetGameName(), message.Action)
	switch message.Action {
	case "join-room":
		//needed only if new room is needed
		rd := GameRoomData{
			IsRandomGame:     false,
			PlayerLimit:      int(message.MessageBody["playerlimit"].(float64)),
			ClientProperties: make(map[string]*ClientProps),
			GameEnded:        make(chan bool),
			Wordslist:        make(map[string]bool),
			Endtime:          1 * 60,
			Rounds:           0,
			TurnAttempted:    make(chan []string),
			HasGameStarted:   false,
			HasGameEnded:     false,
			ClientTurnList:   []*ClientProps{},
		}
		log.Println("JoinRoomAction ", message.Sender, message.MessageBody, room.GetRoomSlugInfo())
		roomname := message.MessageBody["roomname"].(string)
		inroomcnt := len(server.GetClientsInRoom()[roomname])
		log.Println("Players in the room before new added ", roomname, " are", inroomcnt, " and player limit is ", rd.PlayerLimit)
		if inroomcnt < rd.PlayerLimit {
			server.JoinClientRoom(roomname, message.Sender, &rd)
			log.Println("request send to join a room")
		} else {

			clientsinroom := []string{"join-room", roomname, message.Sender}
			r.FailToJoinRoomNotify("room-full", clientsinroom, room, server)
		}
	case "join-random-room":
		rd := GameRoomData{
			IsRandomGame:     true,
			PlayerLimit:      2, //random game only 2
			ClientProperties: make(map[string]*ClientProps),
			GameEnded:        make(chan bool),
			Wordslist:        make(map[string]bool),
			Endtime:          1 * 60,
			Rounds:           0,
			TurnAttempted:    make(chan []string),
			HasGameStarted:   false,
			HasGameEnded:     false,
			ClientTurnList:   []*ClientProps{},
		}
		log.Println("JoinRandomRoomAction ", message.Sender, message.MessageBody, room.GetRoomSlugInfo())
		roomname := message.MessageBody["roomname"].(string)
		if len(r.RandomRooms) > 0 {
			roomname = r.RandomRooms[0]
			inroomcnt := len(server.GetClientsInRoom()[roomname])
			log.Println("Players in the room before new added ", roomname, " are", inroomcnt, " and player limit is ", rd.PlayerLimit)
			if inroomcnt < rd.PlayerLimit {
				server.JoinClientRoom(roomname, message.Sender, &rd)
				log.Println("request send to join a random room")
				if inroomcnt+1 == rd.PlayerLimit { //it means random room reached it limits
					r.RandomRooms = []string{}
				}
			}

			// if len(r.RandomRooms) > 1 {
			// 	r.RandomRooms = r.RandomRooms[1:]
			// } else {
			// 	r.RandomRooms = []string{}
			// }
		} else {
			r.RandomRooms = append(r.RandomRooms, roomname)
			roomname = r.RandomRooms[0]
			inroomcnt := len(server.GetClientsInRoom()[roomname])
			log.Println("Players in the room before new added ", roomname, " are", inroomcnt, " and player limit is ", rd.PlayerLimit)
			if inroomcnt < rd.PlayerLimit {
				server.JoinClientRoom(roomname, message.Sender, &rd)
				log.Println("request send to join a random room")
			}

		}
	}
}

func (r *RoomData) FailToJoinRoomNotify(reason string, clientsinroom []string, room simplysocket.Room, server simplysocket.MeshServer) {
	reasonmsg := ""
	log.Println("Client removed", clientsinroom[2])
	if reason == "room-full" {
		reasonmsg = "Failed to join the room its occupied"
	}
	message := &simplysocket.Message{
		Action: "fail-join-room-notify",
		Target: clientsinroom[2],
		MessageBody: map[string]interface{}{
			"message": reasonmsg,
		},
		Sender:         "bot-of-the-room",
		IsTargetClient: true,
	}

	room.BroadcastMessage(message)
}
func (r *GameRoomData) FailToJoinRoomNotify(reason string, clientsinroom []string, room simplysocket.Room, server simplysocket.MeshServer) {
	reasonmsg := ""
	log.Println("Client removed", clientsinroom[2])
	if reason == "room-full" {
		reasonmsg = "Failed to join the room its occupied"
	}
	message := &simplysocket.Message{
		Action: "fail-join-room-notify",
		Target: clientsinroom[2],
		MessageBody: map[string]interface{}{
			"message": reasonmsg,
		},
		Sender:         "bot-of-the-room",
		IsTargetClient: true,
	}

	room.BroadcastMessage(message)
}
