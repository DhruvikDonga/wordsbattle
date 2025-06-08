package game

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DhruvikDonga/simplysocket"
)

const (
	ChatTimerDuration     = 10 * time.Second
	TurnTimerDuration     = 10 * time.Second
	NextTurnTimerDuration = 11 * time.Second
	PostGameDelay         = 690 * time.Millisecond
	BotSender             = "bot-of-the-room"
	BotName               = "Gawd"
	InitialLetter         = "w"
	ActionByBot           = "message-by-bot"
	// Message action constant
	ActionSetClientName      = "set-client-name"
	ActionStartGame          = "start-the-game"
	ActionRoomSettings       = "room-settings"
	ActionSendMessage        = "send-message"
	ActionAttemptWord        = "attempt-word"
	ActionClientListNotify   = "client-list-notify"
	ActionFailJoinRoom       = "fail-join-room-notify"
	ActionRoomBotGreetings   = "room-bot-greetings"
	ActionKnowYourself       = "know-yourself"
	ActionFoundRandomRoom    = "found-random-room-notify"
	ActionJoinRoomNotify     = "join-room-notify"
	ActionIsRoomMaker        = "is-room-maker"
	ActionRoomBotEndGame     = "room-bot-end-game"
	ActionRoomSettingApplied = "room-setting-applied"
	ActionClientJoinedRoom   = "client-joined-room"
	ActionClientLeftRoom     = "client-left-room"
	ActionJoinRoom           = "join-room"
)

var ColorList = []string{"yellow", "red", "orange", "blue", "purple", "pink", "white"}

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

type ClientsinRoomMessage struct { // we are using this to return list of clients to all clients in room when register unregister happens
	ClientList []*ClientProps `json:"clientlist"` //message
}

func (r *GameRoomData) HandleRoomData(room simplysocket.Room, server simplysocket.MeshServer) {
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
			case ActionClientJoinedRoom:
				if r.HasGameStarted {
					message := &simplysocket.Message{
						Action: ActionFailJoinRoom,
						Target: clientsinroom[2],
						MessageBody: map[string]interface{}{
							"message": "The game in the room has already started",
						},
						Sender:         BotSender,
						IsTargetClient: true,
					}

					room.BroadcastMessage(message)
				} else {
					if r.IsRandomGame {
						r.GotRandomRoom(clientsinroom, room, server)
					}
					r.JoinRoomNotify(clientsinroom, room, server) //to all in the room
					r.KnowTheClient(clientsinroom, room, server)  //to only the client
				}

			case ActionClientLeftRoom:
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
func (r *GameRoomData) handleGameRoommessages(room simplysocket.Room, server simplysocket.MeshServer, message *simplysocket.Message) {
	switch message.Action {
	case ActionSetClientName: //user sends to set his name we will then notify client list
		r.handleSetClientName(message, room, server)
	case ActionStartGame:
		r.HandleStartGameMessage(message.Target, room, server)
	case ActionRoomSettings:
		r.handleRoomSettings(message, room)
	case ActionSendMessage:
		r.broadcastClientMessage(message, room)
	case ActionAttemptWord:
		r.handleAttemptWord(message, room)
	}
}

func (r *GameRoomData) handleSetClientName(message *simplysocket.Message, room simplysocket.Room, server simplysocket.MeshServer) {
	clientsinroom := []string{ActionClientListNotify, message.Target, message.Sender}
	log.Println("Players in the room before new added ", clientsinroom[1], " are", len(r.ClientProperties), " and player limit is ", r.PlayerLimit)
	r.mu.Lock()
	if r.PlayerLimit <= len(r.ClientProperties) {
		r.mu.Unlock()
		clientsinroom := []string{ActionJoinRoom, clientsinroom[1], clientsinroom[2]}
		r.FailToJoinRoomNotify("room-full", clientsinroom, room, server)
		return
	}

	r.ClientProperties[message.Sender] = &ClientProps{Name: message.MessageBody["setname"].(string), Color: ColorList[rand.Intn(len(ColorList))], Score: 0}
	r.mu.Unlock()
	log.Printf("Set client props for %s: %+v", message.Sender, r.ClientProperties[message.Sender])
	r.ClientListNotify(clientsinroom, room, server)
	if r.IsRandomGame && r.PlayerLimit == len(r.ClientProperties) {
		r.HandleStartGameMessage(clientsinroom[1], room, server)
	}
}

func (r *GameRoomData) HandleStartGameMessage(target string, room simplysocket.Room, server simplysocket.MeshServer) {
	log.Println("start the game for room ", target)
	r.HasGameStarted = true
	go func() {
		r.EndTheGameTimer(room, server)
	}()

	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionRoomBotGreetings,
		Target: target,
		MessageBody: map[string]interface{}{
			"message": "welcome",
		},
		Sender:         BotSender,
		IsTargetClient: false,
	})

	time.Sleep(1 * time.Second)

	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionSendMessage,
		Target: target,
		Sender: BotSender,
		MessageBody: map[string]interface{}{
			"message": "Yo this is <b>Bot<small>@room-<small>" + target + "</small></small></b> here . I will be having my 👀 eyes over you if you playing fair, update the score board and assign you new letter",
			"letter":  "",
		},
	})

	time.Sleep(1 * time.Second)
	r.updateClientTurnList()

	r.mu.RLock()
	clist := make([]*ClientProps, len(r.ClientTurnList))
	copy(clist, r.ClientTurnList)
	r.mu.RUnlock()

	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionByBot,
		Target: target,
		Sender: BotSender,
		MessageBody: map[string]interface{}{
			"message":     "Okay giving you 10 seconds ⌛ to communicate with each other your textbox and send button will be activated then will start the match ",
			"timer":       10,
			"clientstats": clist,
			"letter":      "",
		},
	})
	go r.startChatTimer(target, room, server)

}

func (r *GameRoomData) handleRoomSettings(message *simplysocket.Message, room simplysocket.Room) {
	endTime, err1 := strconv.Atoi(message.MessageBody["game_duration"].(string))
	playerLimit, err2 := strconv.Atoi(message.MessageBody["player_limit"].(string))
	if err1 != nil || err2 != nil {
		log.Printf("Error parsing room settings: %v, %v", err1, err2)
		return
	}
	r.mu.Lock()
	r.Endtime = endTime
	r.PlayerLimit = playerLimit
	r.mu.Unlock()
	log.Printf("Room settings updated: PlayerLimit=%d, Endtime=%d", r.PlayerLimit, r.Endtime)
	room.BroadcastMessage(&simplysocket.Message{
		Action:         ActionRoomSettingApplied,
		Target:         message.Target,
		MessageBody:    map[string]interface{}{"message": "Room settings applied successfully Player Limit:- " + message.MessageBody["player_limit"].(string) + " Time duration :-" + message.MessageBody["game_duration"].(string)},
		Sender:         BotSender,
		IsTargetClient: false,
	})
}

func (r *GameRoomData) broadcastClientMessage(message *simplysocket.Message, room simplysocket.Room) {
	r.mu.RLock()
	clientProp := r.ClientProperties[message.Sender]
	r.mu.RUnlock()
	message.MessageBody["useravatar"] = clientProp.Name[:2]
	message.MessageBody["color"] = clientProp.Color
	room.BroadcastMessage(message)
}

func (r *GameRoomData) handleAttemptWord(message *simplysocket.Message, room simplysocket.Room) {
	r.mu.RLock()
	clientProp := r.ClientProperties[message.Sender]
	r.mu.RUnlock()
	message.MessageBody["useravatar"] = clientProp.Name[:2]
	message.MessageBody["color"] = clientProp.Color
	message.Action = "send-message"
	room.BroadcastMessage(message)
	select {
	case r.TurnAttempted <- []string{message.MessageBody["message"].(string), message.Sender}:
	default:
		log.Println("TurnAttempted channel is full, dropping message")
	}
}

func (r *GameRoomData) ClientListNotify(clientsinroom []string, room simplysocket.Room, server simplysocket.MeshServer) {

	if clientsinroom[0] == ActionClientLeftRoom {
		log.Printf("Client %s left room", clientsinroom[2])
		r.removefromclientlist(clientsinroom[2])
		r.mu.Lock()
		delete(r.ClientProperties, clientsinroom[2])
		r.mu.Unlock()

		if r.HasGameStarted && !r.HasGameEnded {
			if len(r.ClientTurnList) == 1 {
				r.endGameDueToPlayerLeft(room)
				return
			} else {
				r.updateClientTurnList()
			}
		}
	} else {
		r.updateClientTurnList()
	}

	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionClientListNotify,
		Target: clientsinroom[1],
		MessageBody: map[string]interface{}{
			"clientsinroomessage": r.ClientTurnList,
		},
		Sender:         BotName,
		IsTargetClient: false,
	})
}

func (r *GameRoomData) endGameDueToPlayerLeft(room simplysocket.Room) {
	log.Printf("Game in room %s ended due to 1 player left", r.Slug)
	r.HasGameEnded = true
	r.GameEnded <- true
	room.BroadcastMessage(&simplysocket.Message{
		Action:      ActionSendMessage,
		MessageBody: map[string]interface{}{"message": "Game ended successfully due to rest players left🍾 "},
		Target:      room.GetRoomSlugInfo(),
		Sender:      BotSender,
	})
	time.Sleep(PostGameDelay)
	message := "Game ended successfully due to you are only left to play 🍾 "
	r.sendGameEndStats(room, message)
}

func (r *GameRoomData) sendGameEndStats(room simplysocket.Room, message string) {
	clientList := r.ClientTurnList

	sort.Slice(clientList[:], func(i, j int) bool {
		return clientList[i].Score > clientList[j].Score
	})
	wordlist := []string{}
	for words := range r.Wordslist {
		wordlist = append(wordlist, words)
	}
	room.BroadcastMessage(&simplysocket.Message{
		Action:      ActionRoomBotEndGame,
		MessageBody: map[string]interface{}{"message": message, "client_list": clientList, "word_list": wordlist},
		Target:      room.GetRoomSlugInfo(),
		Sender:      BotSender,
	})
}

func (r *GameRoomData) updateClientTurnList() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ret := make([]*ClientProps, 0, len(r.ClientProperties))
	for slug, props := range r.ClientProperties {
		ret = append(ret, &ClientProps{
			Color: props.Color,
			Name:  props.Name,
			Score: props.Score,
			Slug:  slug,
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Slug < ret[j].Slug
	})

	r.ClientTurnList = ret
}

func (r *GameRoomData) startChatTimer(target string, room simplysocket.Room, server simplysocket.MeshServer) {
	firsturn := time.After(ChatTimerDuration)
	for {
		select {
		case <-firsturn:
			log.Println("Chat time ended")
			r.updateClientTurnList()
			r.mu.Lock()
			clist := r.ClientTurnList
			r.WhichClientTurn = clist[0]
			r.Letter = InitialLetter
			r.mu.Unlock()
			room.BroadcastMessage(&simplysocket.Message{
				Action: ActionByBot,
				Target: target,
				Sender: BotSender,
				MessageBody: map[string]interface{}{
					"message":         "Cool Cool , Enough of talking start the show <br><b>" + clist[0].Name + "<small>@" + clist[0].Slug + "</small></b> <br> start with letter <b>W</b> <br>Time starts now 10 seconds ⌛",
					"clientstats":     clist,
					"letter":          r.Letter,
					"whichclientturn": r.WhichClientTurn,
					"timer":           10,
				},
			})
			r.TurnTheGameTimer(room, server)
			return
		case <-room.RoomStopped():
			log.Println("chat only timer routine stopped cause room is stopped")
			return
		}
	}
}

func (r *GameRoomData) KnowTheClient(clientsinroom []string, room simplysocket.Room, server simplysocket.MeshServer) {
	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionKnowYourself,
		Target: clientsinroom[2],
		MessageBody: map[string]interface{}{
			"sender": clientsinroom[2],
		},
		Sender:         BotSender,
		IsTargetClient: true,
	})
}

func (r *GameRoomData) GotRandomRoom(clientsinroom []string, room simplysocket.Room, server simplysocket.MeshServer) {
	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionFoundRandomRoom,
		Target: clientsinroom[2],
		MessageBody: map[string]interface{}{
			"roomname": clientsinroom[1],
		},
		Sender:         BotSender,
		IsTargetClient: true,
	})
}

func (r *GameRoomData) JoinRoomNotify(clientsinroom []string, room simplysocket.Room, server simplysocket.MeshServer) {

	room.BroadcastMessage(&simplysocket.Message{
		Action: ActionJoinRoomNotify,
		Target: clientsinroom[1],
		MessageBody: map[string]interface{}{
			"newmessage": fmt.Sprintf("%s joined the room cowgame by mesh", clientsinroom[2]),
		},
		Sender:         clientsinroom[2],
		IsTargetClient: false,
	})

	if clientsinroom[2] == room.GetRoomMakerInfo() {
		room.BroadcastMessage(&simplysocket.Message{
			Action:         ActionIsRoomMaker,
			Target:         clientsinroom[2],
			MessageBody:    map[string]interface{}{"message": clientsinroom[2]},
			Sender:         BotSender,
			IsTargetClient: true,
		})
	}
}

func (r *GameRoomData) TurnTheGameTimer(room simplysocket.Room, server simplysocket.MeshServer) {

	endtime := time.After(TurnTimerDuration)
	go func() {
		for {
			select {
			case <-endtime:
				if !r.processTurnTimeout(room) {
					return
				}
				endtime = time.After(NextTurnTimerDuration)

			case wordguessedbyclient := <-r.TurnAttempted:
				if !r.processTurnAttempt(wordguessedbyclient, room) {
					return
				}
				endtime = time.After(NextTurnTimerDuration)

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

func (r *GameRoomData) processTurnTimeout(room simplysocket.Room) bool {
	r.Rounds += 1
	log.Println("the turn in room ", room.GetRoomSlugInfo(), " ended round", r.Rounds)
	room.BroadcastMessage(&simplysocket.Message{
		Action:      ActionSendMessage,
		MessageBody: map[string]interface{}{"message": "Game turn ended successfully 🍾 "},
		Target:      room.GetRoomSlugInfo(),
		Sender:      BotSender,
	})

	r.mu.RLock()
	currentplayer := r.WhichClientTurn
	r.mu.RUnlock()

	if currentplayer == nil {
		return false
	}

	nextplayer := r.getnextplayer(currentplayer)
	if nextplayer == nil {
		return false
	}

	r.mu.Lock()
	r.WhichClientTurn = nextplayer
	r.mu.Unlock()

	resmessage := fmt.Sprintf("Word not guessed by, <b>%v<small>@%v</small></b> times up <br> now <b>%v<small>@%v</small></b> start with letter <b>%v</b> <br>Time starts now 10 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
	time.Sleep(1 * time.Second)
	room.BroadcastMessage(&simplysocket.Message{
		Target:      room.GetRoomSlugInfo(),
		MessageBody: map[string]interface{}{"message": resmessage, "whichclientturn": nextplayer, "clientstats": r.ClientTurnList, "letter": r.Letter, "timer": 11},
		Action:      ActionByBot,
		Sender:      BotSender,
	})
	return true
}

func (r *GameRoomData) processTurnAttempt(wordguessedbyclient []string, room simplysocket.Room) bool {
	room.BroadcastMessage(&simplysocket.Message{
		Action:      ActionSendMessage,
		MessageBody: map[string]interface{}{"message": "Game turn ended successfully 🍾 "},
		Target:      room.GetRoomSlugInfo(),
		Sender:      BotSender,
	})

	r.Rounds += 1
	guessedword := strings.ToLower(wordguessedbyclient[0])
	log.Println("user ", wordguessedbyclient[1], " attempted the turn", guessedword, " its round", r.Rounds, "letter was", r.Letter)

	status := MatchWord(guessedword, r.Wordslist, r.Letter[0])

	r.mu.RLock()
	currentplayer := r.WhichClientTurn
	r.mu.RUnlock()

	if currentplayer == nil {
		return false
	}

	nextplayer := r.getnextplayer(currentplayer)
	if nextplayer == nil {
		return false
	}

	resmessage := r.generateResponseMessage(status, guessedword, currentplayer, nextplayer)

	r.mu.Lock()
	r.WhichClientTurn = nextplayer
	r.mu.Unlock()

	time.Sleep(1 * time.Second)
	room.BroadcastMessage(&simplysocket.Message{
		Action:      ActionByBot,
		MessageBody: map[string]interface{}{"message": resmessage, "whichclientturn": nextplayer, "clientstats": r.ClientTurnList, "letter": r.Letter, "timer": 11},
		Target:      room.GetRoomSlugInfo(),
		Sender:      BotSender,
	})
	return true
}

func (r *GameRoomData) generateResponseMessage(status, guessedword string, currentplayer, nextplayer *ClientProps) string {
	switch status {
	case WordCorrect:
		r.Letter = string(guessedword[len(guessedword)-1]) //correct then new letter
		r.mu.Lock()
		r.WhichClientTurn.Score += 1
		r.Wordslist[guessedword] = true //add it to our word list
		r.mu.Unlock()
		return fmt.Sprintf("Bravo correct word guessed by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
	case WrongLetter:
		return fmt.Sprintf("Word starts with wrong letter <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
	case NoSuchWord:
		return fmt.Sprintf("No such word exisists in our dictionary guessed by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
	case WordReused:
		return fmt.Sprintf("This word is already guessed  <b>%v<small>@%v</small></b> so not helpfull now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
	default:
		return fmt.Sprintf("Unknown error by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 11 seconds ⌛",
			currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, r.Letter)
	}
}

func (r *GameRoomData) EndTheGameTimer(room simplysocket.Room, server simplysocket.MeshServer) {

	endtime := time.After(time.Duration(r.Endtime) * time.Second)
	for {
		select {
		case <-endtime:
			log.Println("the game in room ", room.GetRoomSlugInfo(), " ended")
			r.HasGameEnded = true
			r.GameEnded <- true
			room.BroadcastMessage(&simplysocket.Message{
				Action:      ActionSendMessage,
				MessageBody: map[string]interface{}{"message": "Game ended successfully 🍾 "},
				Target:      room.GetRoomSlugInfo(),
				Sender:      BotSender,
			})
			time.Sleep(PostGameDelay)
			message := "Game ended successfully 🍾 "
			r.sendGameEndStats(room, message)
			return
		case <-room.RoomStopped():
			log.Println("start timer routine stopped cause room is stopped")
			return
		}
	}
}

func (r *GameRoomData) getnextplayer(current *ClientProps) *ClientProps {
	for i, client := range r.ClientTurnList {
		if client.Slug == current.Slug {
			return r.ClientTurnList[(i+1)%len(r.ClientTurnList)]
		}
	}
	return nil
}

func (r *GameRoomData) removefromclientlist(slug string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, client := range r.ClientTurnList {
		if client.Slug == slug {
			log.Println("client left room triggered removefromclientlist", i, true, slug)
			r.ClientTurnList = append(r.ClientTurnList[:i], r.ClientTurnList[i+1:]...)
			return
		}
	}

	log.Println("client left room triggered removefromclientlist - not found", -1, false, slug)
}
