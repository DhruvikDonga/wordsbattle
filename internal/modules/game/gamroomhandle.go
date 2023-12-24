package game

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
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
	ClientTurnList   []*ClientProps
	Rounds           int
	Endtime          int //time till which game is to be played
	WhichClientTurn  *ClientProps
	GameEnded        chan bool
	Letter           string
	TurnAttempted    chan []string //use to close the ticker if user has responded before the timer ends4
	HasGameStarted   bool
	HasGameEnded     bool
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
				if r.HasGameStarted {
					message := &gomeshstream.Message{
						Action: "fail-join-room-notify",
						Target: clientsinroom[2],
						MessageBody: map[string]interface{}{
							"message": "The game in the room has already started",
						},
						Sender:         "bot-of-the-room",
						IsTargetClient: true,
					}

					server.BroadcastMessage(message)
				} else {
					r.JoinRoomNotify(clientsinroom, room, server) //to all in the room
					r.KnowTheClient(clientsinroom, room, server)  //to only the client
				}

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

	case "send-message":
		message.MessageBody["useravatar"] = r.ClientProperties[message.Sender].Name[:2]
		message.MessageBody["color"] = r.ClientProperties[message.Sender].Color
		room.BroadcastMessage(message)

	case "attempt-word":
		message.MessageBody["useravatar"] = r.ClientProperties[message.Sender].Name[:2]
		message.MessageBody["color"] = r.ClientProperties[message.Sender].Color
		message.Action = "send-message"
		room.BroadcastMessage(message)
		select {
		default:
		case r.TurnAttempted <- []string{message.MessageBody["message"].(string), message.Sender}:

		}

	}

}

type ClientsinRoomMessage struct { // we are using this to return list of clients to all clients in room when register unregister happens
	ClientList []*ClientProps `json:"clientlist"` //message
}

func (r *GameRoomData) ClientListNotify(clientsinroom []string, room gomeshstream.Room, server gomeshstream.MeshServer) {
	ret := []*ClientProps{}

	if clientsinroom[0] == "client-left-room" {
		log.Println("client left room triggered")
		r.removefromclientlist(clientsinroom[2])

		r.mu.Lock()
		delete(r.ClientProperties, clientsinroom[2])

		r.mu.Unlock()

	} else {
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
		r.ClientTurnList = ret
	}
	message := &gomeshstream.Message{
		Action: "client-list-notify",
		Target: clientsinroom[1],
		MessageBody: map[string]interface{}{
			"clientsinroomessage": r.ClientTurnList,
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
	r.HasGameStarted = true
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
			"letter":  "",
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
	r.ClientTurnList = clist
	chattimemessage := &gomeshstream.Message{
		Action: "message-by-bot",
		Target: target,
		Sender: "bot-of-the-room",
		MessageBody: map[string]interface{}{
			"message":     "Okay giving you 10 seconds ⌛ to communicate with each other your textbox and send button will be activated then will start the match ",
			"timer":       10,
			"clientstats": clist,
			"letter":      "",
		},
	}
	room.BroadcastMessage(chattimemessage)
	go func() {
		firsturn := time.After(10 * time.Second)
		for {
			select {
			case <-firsturn:
				log.Println("chattime ends")
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
				r.ClientTurnList = clist
				r.WhichClientTurn = clist[0]
				r.Letter = "w"
				chattimeendmessage := &gomeshstream.Message{
					Action: "message-by-bot",
					Target: target,
					Sender: "bot-of-the-room",
					MessageBody: map[string]interface{}{
						"message":         "Cool Cool , Enough of talking start the show <br><b>" + clist[0].Name + "<small>@" + clist[0].Slug + "</small></b> <br> start with letter <b>W</b> <br>Time starts now 10 seconds ⌛",
						"clientstats":     clist,
						"letter":          r.Letter,
						"whichclientturn": r.WhichClientTurn,
						"timer":           10,
					},
				}
				room.BroadcastMessage(chattimeendmessage)
				r.TurnTheGameTimer(room, server)
				return
			case <-room.RoomStopped():
				log.Println("chat only timer routine stopped cause room is stopped")
				return
			}
		}
	}()

}

func (r *GameRoomData) TurnTheGameTimer(room gomeshstream.Room, server gomeshstream.MeshServer) {

	endtime := time.After(10 * time.Second)
	go func() {
		for {
			select {
			case <-endtime:
				r.Rounds += 1
				log.Println("the turn in room ", room.GetRoomSlugInfo(), " ended round", r.Rounds)
				message := &gomeshstream.Message{
					Action:      "send-message",
					MessageBody: map[string]interface{}{"message": "Game turn ended successfully 🍾 "},
					Target:      room.GetRoomSlugInfo(),
					Sender:      "bot-of-the-room",
				}
				room.BroadcastMessage(message)

				r.mu.RLock()
				currentplayer := r.WhichClientTurn
				r.mu.RUnlock()
				nextplayer := r.getnextplayer(currentplayer)
				r.WhichClientTurn = nextplayer
				resmessage := fmt.Sprintf("Word not guessed by, <b>%v<small>@%v</small></b> times up <br> now <b>%v<small>@%v</small></b> start with letter <b>%v</b> <br>Time starts now 10 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)

				sendmessage := &gomeshstream.Message{
					Target:      room.GetRoomSlugInfo(),
					MessageBody: map[string]interface{}{"message": resmessage, "whichclientturn": nextplayer, "clientstats": r.ClientTurnList, "letter": r.Letter, "timer": 11},
					Action:      "message-by-bot",
					Sender:      "bot-of-the-room",
				}
				time.Sleep(1 * time.Second)
				room.BroadcastMessage(sendmessage)
				endtime = time.After(11 * time.Second)

			case wordguessedbyclient := <-r.TurnAttempted:
				message := &gomeshstream.Message{
					Action:      "send-message",
					MessageBody: map[string]interface{}{"message": "Game turn ended successfully 🍾 "},
					Target:      room.GetRoomSlugInfo(),
					Sender:      "bot-of-the-room",
				}
				room.BroadcastMessage(message)
				r.Rounds += 1
				guessedword := strings.ToLower(wordguessedbyclient[0])
				log.Println("user ", wordguessedbyclient[1], " attempted the turn", guessedword, " its round", r.Rounds, "letter was", r.Letter)
				status := MatchWord(guessedword, r.Wordslist, r.Letter[0])
				r.mu.RLock()
				currentplayer := r.WhichClientTurn
				r.mu.RUnlock()
				nextplayer := r.getnextplayer(currentplayer)
				resmessage := ""
				if status == "word-correct" {
					r.Letter = string(guessedword[len(guessedword)-1]) //correct then new letter
					resmessage = fmt.Sprintf("Bravo correct word guessed by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
					r.mu.Lock()
					r.WhichClientTurn.Score += 1
					r.mu.Unlock()
					r.Wordslist[guessedword] = true //add it to our word list

				} else if status == "wrong-letter" {
					resmessage = fmt.Sprintf("Word starts with wrong letter <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
				} else if status == "no-such-word" {
					resmessage = fmt.Sprintf("No such word exisists in our dictionary guessed by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
				} else if status == "word-reused" {
					resmessage = fmt.Sprintf("This word is already guessed  <b>%v<small>@%v</small></b> so not helpfull now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
				}
				r.WhichClientTurn = nextplayer

				message = &gomeshstream.Message{
					Action:      "message-by-bot",
					MessageBody: map[string]interface{}{"message": resmessage, "whichclientturn": nextplayer, "clientstats": r.ClientTurnList, "letter": r.Letter, "timer": 11},
					Target:      room.GetRoomSlugInfo(),
					Sender:      "bot-of-the-room",
				}
				time.Sleep(1 * time.Second)
				room.BroadcastMessage(message)

				endtime = time.After(11 * time.Second)
			case <-room.RoomStopped():
				log.Println("start timer routine stopped cause room is stopped")
				return

			case <-r.GameEnded:
				log.Println("game timer routine stopped cause game ended")
				return
			}
		}
	}()
}

func (r *GameRoomData) EndTheGameTimer(room gomeshstream.Room, server gomeshstream.MeshServer) {

	endtime := time.After(time.Duration(r.Endtime) * time.Second)
	for {
		select {
		case <-endtime:
			log.Println("the game in room ", room.GetRoomSlugInfo(), " ended")
			r.HasGameEnded = true
			r.GameEnded <- true
			message := &gomeshstream.Message{
				Action:      "send-message",
				MessageBody: map[string]interface{}{"message": "Game ended successfully 🍾 "},
				Target:      room.GetRoomSlugInfo(),
				Sender:      "bot-of-the-room",
			}
			room.BroadcastMessage(message)
			time.Sleep(690 * time.Millisecond)
			clist := r.ClientTurnList

			sort.Slice(clist[:], func(i, j int) bool {
				return clist[i].Score > clist[j].Score
			})
			wordlist := []string{}
			for words := range r.Wordslist {
				wordlist = append(wordlist, words)
			}
			messagestats := &gomeshstream.Message{
				Action:      "room-bot-end-game",
				MessageBody: map[string]interface{}{"message": "Game ended successfully 🍾 ", "client_list": clist, "word_list": wordlist},
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

func (r *GameRoomData) getnextplayer(currentplayer *ClientProps) *ClientProps {
	index := 0
	clientlist := r.ClientTurnList
	for key, client := range clientlist {
		if client.Slug == currentplayer.Slug {
			index = key
		}
	}
	if index+1 == len(clientlist) { //last key
		index = 0 //first client
	} else {
		index += 1 //next client
	}
	return clientlist[index]
}

func (r *GameRoomData) removefromclientlist(removedplayerslug string) {
	r.mu.Lock()
	index := 0
	flg := false

	for key, client := range r.ClientTurnList {
		if client.Slug == removedplayerslug {
			index = key
			flg = true
			break
		}
	}
	log.Println("client left room triggered remmovefromclientlist", index, flg, removedplayerslug)

	if flg {
		if len(r.ClientTurnList) > 1 {
			r.ClientTurnList = append(r.ClientTurnList[:index], r.ClientTurnList[index+1:]...)
		} else {
			r.ClientTurnList = []*ClientProps{}
		}
	}
	r.mu.Unlock()
}
