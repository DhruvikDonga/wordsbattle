package game

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/DhruvikDonga/wordsbattle/pkg/gomeshstream"
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
	Endtime          int //time till which game is to be played
}
type ClientProps struct { //this can depend on a inroom basis so it changes
	Color string `json:"color"`
	Name  string `json:"name"`
	Score int    `json:"score"`
	Slug  string `json:"slug"`
}

func (r *GameRoomData) HandleRoomData(room gomeshstream.Room, server gomeshstream.MeshServer) {
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
func (r *GameRoomData) handleGameRoommessages(room gomeshstream.Room, server gomeshstream.MeshServer, message *gomeshstream.Message) {
	switch message.Action {
	case "set-client-name": //user sends to set his name we will then notify client list
		colors := []string{"yellow", "red", "orange", "blue", "purple", "pink", "white"}
		color := colors[rand.Intn(len(colors))]
		r.mu.Lock()
		r.ClientProperties[message.Sender] = &ClientProps{Name: message.MessageBody["setname"].(string), Color: color, Score: 0}
		r.mu.Unlock()
		log.Printf("Set client props: %+v\n ", r.ClientProperties[message.Sender])
		clientsinroom := []string{"client-list-notify", message.Target, message.Sender}
		r.ClientListNotify(clientsinroom, room, server)

	case "start-the-game":
		r.HandleStartGameMessage(message.Sender, message.Target, room, server)
		go func() {
			r.EndTheGameTimer(room, server)
		}()
	}

}

type ClientsinRoomMessage struct { // we are using this to return list of clients to all clients in room when register unregister happens
	ClientList []*ClientProps `json:"clientlist"` //message
}

func (r *GameRoomData) ClientListNotify(clientsinroom []string, room gomeshstream.Room, server gomeshstream.MeshServer) {
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
	sort.Slice(ret[:], func(i, j int) bool {
		return ret[i].Slug < ret[j].Slug
	})
	message := &gomeshstream.Message{
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

func (r *GameRoomData) KnowTheClient(clientsinroom []string, room gomeshstream.Room, server gomeshstream.MeshServer) {
	message := &gomeshstream.Message{
		Action: "know-yourself",
		Target: clientsinroom[2],
		MessageBody: map[string]interface{}{
			"sender": clientsinroom[2],
		},
		Sender:         "bot-of-the-room",
		IsTargetClient: true,
	}

	server.BroadcastMessage(message)
}

func (r *GameRoomData) JoinRoomNotify(clientsinroom []string, room gomeshstream.Room, server gomeshstream.MeshServer) {
	message := &gomeshstream.Message{
		Action: "join-room-notify",
		Target: clientsinroom[1],
		MessageBody: map[string]interface{}{
			"newmessage": fmt.Sprintf("%s joined the room cowgame by mesh", clientsinroom[2]),
		},
		Sender:         clientsinroom[2],
		IsTargetClient: false,
	}

	server.BroadcastMessage(message)

	if clientsinroom[2] == room.GetRoomMakerInfo() {
		roommakermessage := &gomeshstream.Message{
			Action:         "is-room-maker",
			Target:         clientsinroom[2],
			MessageBody:    map[string]interface{}{"message": clientsinroom[2]},
			Sender:         "bot-of-the-room",
			IsTargetClient: true,
		}
		server.BroadcastMessage(roommakermessage)

	}
}

func (r *GameRoomData) HandleStartGameMessage(sender, target string, room gomeshstream.Room, server gomeshstream.MeshServer) {
	log.Println("start the game for room ", target)
	message := &gomeshstream.Message{
		Action: "room-bot-greetings",
		Target: target,
		MessageBody: map[string]interface{}{
			"message": "welcome",
		},
		Sender:         "bot-of-the-room",
		IsTargetClient: false,
	}

	room.BroadcastMessage(message)

	time.Sleep(1 * time.Second)
	botgreetingsmessage := &gomeshstream.Message{
		Action: "send-message",
		Target: target,
		Sender: "bot-of-the-room",
		MessageBody: map[string]interface{}{
			"message": "Yo this is <b>Bot<small>@room-<small>" + target + "</small></small></b> here . I will be having my 👀 eyes over you if you playing fair, update the score board and assign you new letter",
		},
	}
	room.BroadcastMessage(botgreetingsmessage)

	time.Sleep(1 * time.Second)
	clist := []*ClientProps{}
	r.mu.RLock()
	for slug, props := range r.ClientProperties {
		temp := &ClientProps{
			Color: props.Color,
			Name:  props.Name,
			Score: props.Score,
			Slug:  slug,
		}
		clist = append(clist, temp)
	}
	r.mu.RUnlock()
	sort.Slice(clist[:], func(i, j int) bool {
		return clist[i].Slug < clist[j].Slug
	})
	chattimemessage := &gomeshstream.Message{
		Action: "message-by-bot",
		Target: target,
		Sender: "bot-of-the-room",
		MessageBody: map[string]interface{}{
			"message":     "Okay giving you 10 seconds ⌛ to communicate with each other your textbox and send button will be activated then will start the match ",
			"timer":       10,
			"clientstats": clist,
		},
	}
	room.BroadcastMessage(chattimemessage)

}

func (r *GameRoomData) EndTheGameTimer(room gomeshstream.Room, server gomeshstream.MeshServer) {

	endtime := time.After(time.Duration(r.Endtime) * time.Second)
	for {
		select {
		case <-endtime:
			log.Println("the game in room ", room.GetRoomSlugInfo(), " ended")
			message := &gomeshstream.Message{
				Action:      "send-message",
				MessageBody: map[string]interface{}{"message": "Game ended successfully 🍾 "},
				Target:      room.GetRoomSlugInfo(),
				Sender:      "bot-of-the-room",
			}
			room.BroadcastMessage(message)
			time.Sleep(690 * time.Millisecond)
			clist := []*ClientProps{}
			r.mu.RLock()
			for slug, props := range r.ClientProperties {
				temp := &ClientProps{
					Color: props.Color,
					Name:  props.Name,
					Score: props.Score,
					Slug:  slug,
				}
				clist = append(clist, temp)
			}
			r.mu.RUnlock()
			sort.Slice(clist[:], func(i, j int) bool {
				return clist[i].Score > clist[j].Score
			})
			words := []string{}
			messagestats := &gomeshstream.Message{
				Action:      "room-bot-end-game",
				MessageBody: map[string]interface{}{"message": "Game ended successfully 🍾 ", "client_list": clist, "word_list": words},
				Target:      room.GetRoomSlugInfo(),
				Sender:      "bot-of-the-room",
			}
			room.BroadcastMessage(messagestats)
			return
		case <-room.RoomStopped():
			log.Println("start timer routine stopped cause room is stopped")
			return
		}
	}
}
