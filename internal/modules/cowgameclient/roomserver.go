package cowgameclient

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// room server is a goroutine created by the chatServer when "join-room" message is catched by read-pump it either joins the user to exisiting room or create a new room
// room will handle new rooms associated clients register, unregister a client into it and also broadcast to clients

type GameRoomMetadata struct { //this can change as per game and it should be initialized when game is started
	gamestarted     bool
	gameendtime     time.Time
	gameended       bool
	gamechatevent   bool            //this is trial purpose if this event is false will have to do it
	gamechattime    time.Time       //this is chat time in cases
	gameturn        bool            //if set to true user has played his turn
	gameturnskip    bool            //if set to true user skipped his turn might be he left the server
	gameturntime    time.Time       //this is recursive will set to null then start again can be different in cases
	wordslist       map[string]bool //word and meaning
	clientlist      []*Client
	whichclientturn *Client
	letter          string //[0]
	wordguessed     string
	rounds          int
}

type Room struct {
	name                    string
	clients                 map[*Client]bool
	wsServer                *LobbyServer //keep reference of webserver to every client
	register                chan *Client
	unregister              chan *Client
	broadcast               chan *Message //message send in a room
	broadcastToClientinRoom chan *Message //message to a client
	broadcastbyBot          chan *RoomBotGameMessage
	stoproom                chan bool //end the gameserver routine
	randomgame              bool      //true then it is random game 2 players only once player connect then start the game automatic no lobby system
	roommaker               *Client   //room maker will get a seperate message which will make him capable to start the game
	gamemetadata            *GameRoomMetadata
	playerlimit             int
}

func NewRoom(name string, gameServer *LobbyServer, roommaker *Client) *Room {
	return &Room{
		name:                    name,
		wsServer:                gameServer,
		clients:                 make(map[*Client]bool),
		register:                make(chan *Client),
		unregister:              make(chan *Client),
		broadcast:               make(chan *Message), //unbuffered channel unlike of send of client cause it will recieve only when readpump sends in it else it will block
		broadcastToClientinRoom: make(chan *Message),
		broadcastbyBot:          make(chan *RoomBotGameMessage),
		stoproom:                make(chan bool),
		//randomgame:               true, //set it during create room
		roommaker:    roommaker,
		gamemetadata: &GameRoomMetadata{gamestarted: false, gamechatevent: false, letter: "", wordguessed: "", gameturn: false, gameturnskip: false, wordslist: make(map[string]bool), rounds: 0}, //game not started in room
	}
}

// Run websocket server accepting various requests
func (room *Room) RunRoom() {
	ticker := time.NewTicker(5 * time.Millisecond) //need update for room every 10 millisecond
	for {
		select {
		case client := <-room.register:
			log.Println("Client name;-", client.Name)
			if room.randomgame == true { //notify the roomname to client
				message := &Message{
					Action:  FoundRandomRoomNotification,
					Message: fmt.Sprintf(room.name),
				}
				client.send <- message.encode()
			}
			room.registerClientinRoom(client) //add the client
			room.notifyClientJoined(client)   // notify to clients in room

			//notify to only that client registered who has created room
			if room.roommaker.Slug == client.Slug {
				log.Println("sending roommaker")
				roommakermessage := &Message{
					Action:  RoomMakerNotification,
					Target:  room.name,
					Message: client.Slug,
				}
				client.send <- roommakermessage.encode()
			}
			//clients in room notification
			rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
			colors := []string{"yellow", "red", "orange", "blue", "purple", "pink", "white"}
			color := colors[rand.Intn(len(colors))]
			client.Gamemetadata.Color = color //initial color
			client.Gamemetadata.Score = 0     //initial score in room
			var clientlist []*Client
			for key := range room.clients {
				clientlist = append(clientlist, key)
			}
			sort.Slice(clientlist[:], func(i, j int) bool {
				return clientlist[i].Slug < clientlist[j].Slug
			})
			message := &ClientsinRoomMessage{
				Action:     ClientListNotification,
				Target:     room.name,
				ClientList: clientlist,
			}
			room.broadcastToClientsInRoom(message.encode())
			knowyourself := &Message{
				Action: KnowYourUser,
				Target: room.name,
				Sender: client,
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: knowyourself}
			if room.randomgame == true && len(room.clients) == room.playerlimit { //random game so if we get player limit reached we will start the game from server side
				botnotify := &Message{ //imp will change ui
					Action: StartGameNotification,
					Target: room.name,
					Sender: &Client{Name: "bot-of-the-room"},
				}
				time.Sleep(1 * time.Second)
				room.broadcastToClientsInRoom(botnotify.encode())
				room.gamemetadata.gamestarted = true //game started
			}
		case client := <-room.unregister:
			if room.gamemetadata.gamestarted == true && room.gamemetadata.gameended == false {
				if room.gamemetadata.whichclientturn != nil {
					if client == room.gamemetadata.whichclientturn { //room leaver had a turn
						log.Println("CLinet had turn he left", client.Name, client.Slug) //better action here
						room.gamemetadata.gameturnskip = true                            //skip the turn and let see
					}
				}
				room.unregisterClientinRoom(client) //remove the client
				var clientlist []*Client
				for key := range room.clients {
					clientlist = append(clientlist, key)
				}
				sort.Slice(clientlist[:], func(i, j int) bool {
					return clientlist[i].Slug < clientlist[j].Slug
				})
				room.gamemetadata.clientlist = clientlist
			} else {

				room.unregisterClientinRoom(client) //remove the client
			}
			//clients in room notification
			var clientlist []*Client
			for key := range room.clients {
				clientlist = append(clientlist, key)
			}
			message := &ClientsinRoomMessage{
				Action:     ClientListNotification,
				Target:     room.name,
				ClientList: clientlist,
			}
			room.broadcastToClientsInRoom(message.encode())
			client.ingame = false // is not in any game
			client.room = &Room{} // empty room
			client.Gamemetadata = &ClientGameMetadata{}
			log.Println("Check if room is to be deleted")
			if len(room.clients) == 0 {
				log.Println("room shutdown", room.name)
				room.wsServer.deleteRoom(room)
				return
			}
		case message := <-room.broadcast:
			room.broadcastToClientsInRoom(message.encode()) //broadcast the message from readpump to a room clients only
		case message := <-room.broadcastbyBot:
			room.broadcastToClientsInRoom(message.encode())
		case <-room.stoproom:
			log.Println("room shutdown", room.name)
			room.wsServer.deleteRoom(room)
			return
		case <-ticker.C:
			if room.gamemetadata.gamestarted == true && room.gamemetadata.gameended == false {
				if room.gamemetadata.gameendtime.IsZero() { //game time set
					room.gamemetadata.gameendtime = time.Now().Add(3 * 60 * time.Second) //chat time included
				}

				if !room.gamemetadata.gameendtime.After(time.Now()) { // game ended

					room.broadcastgamestats()
				}
				if room.gamemetadata.gamechatevent == false { //grettings and chat
					room.gamemetadata.gamechatevent = true
					log.Println("start the game with chat")
					time.Sleep(1 * time.Second)
					botgreetingsmessage := &Message{
						Action:  SendMessageAction,
						Target:  room.name,
						Sender:  &Client{Name: "bot-of-the-room"},
						Message: "Yo this is <b>Bot<small>@room-<small>" + room.name + "</small></small></b> here . I will be having my 👀 eyes over you if you playing fair, update the score board and assign you new letter",
					}
					room.broadcastToClientsInRoom(botgreetingsmessage.encode())
					var clientlist []*Client
					for key := range room.clients {
						clientlist = append(clientlist, key)
					}
					sort.Slice(clientlist[:], func(i, j int) bool {
						return clientlist[i].Slug < clientlist[j].Slug
					})
					room.gamemetadata.clientlist = clientlist
					time.Sleep(3 * time.Second)
					botmessage := &RoomBotGameMessage{
						Target:     room.name,
						Message:    "Okay giving you 30 seconds ⌛ to communicate with each other your textbox and send button will be activated then will start the match ",
						Action:     MessageByBot,
						Timer:      30, //15seconds
						ClientList: clientlist,
					}
					room.broadcastToClientsInRoom(botmessage.encode())
					room.gamemetadata.gamechattime = time.Now().Add(30 * time.Second)
				}

				if !room.gamemetadata.gamechattime.After(time.Now()) && room.gamemetadata.rounds == 0 { //chat event ended and no letter means first round
					log.Println("Chat time ends")
					var clientlist []*Client
					for key := range room.clients {
						clientlist = append(clientlist, key)
					}
					sort.Slice(clientlist[:], func(i, j int) bool {
						return clientlist[i].Slug < clientlist[j].Slug
					})
					room.gamemetadata.letter = "W"
					room.gamemetadata.whichclientturn = room.gamemetadata.clientlist[0]
					time.Sleep(1 * time.Second)
					firstwordbotmessage := &RoomBotGameMessage{
						Target:          room.name,
						Message:         fmt.Sprintf("Cool Cool , Enough of talking start the show <b>%v<small>@%v</small></b> <br> start with letter <b>W</b> <br>Time starts now 18 seconds ⌛", room.gamemetadata.clientlist[0].Name, room.gamemetadata.clientlist[0].Slug),
						Action:          MessageByBot,
						WhichClientTurn: room.gamemetadata.whichclientturn, //first one lets go
						ClientList:      room.gamemetadata.clientlist,
						Letter:          room.gamemetadata.letter,
						Timer:           18,
					}
					room.broadcastToClientsInRoom(firstwordbotmessage.encode())
					room.gamemetadata.gameturntime = time.Now().Add(20 * time.Second)
					log.Println(room.gamemetadata.gameturntime)
					room.gamemetadata.gameturn = false
					room.gamemetadata.rounds = room.gamemetadata.rounds + 1
				}
				if !room.gamemetadata.gameturntime.After(time.Now()) && room.gamemetadata.rounds > 0 { //game turn time ends
					log.Println("Player turn ended")
					room.gamemetadata.rounds = room.gamemetadata.rounds + 1
					room.gamemetadata.gameturnskip = false
					room.gamemetadata.gameturn = false
					room.gamemetadata.gameturntime = time.Now().Add(18 * time.Second)
					room.reportuserturnend()
				}
				if room.gamemetadata.gameturntime.After(time.Now()) && room.gamemetadata.gameturn == true && room.gamemetadata.rounds > 0 { //client made a turn before turn time ends
					log.Println("Word guessed by user", room.gamemetadata.letter)
					room.gamemetadata.rounds = room.gamemetadata.rounds + 1
					room.gamemetadata.gameturn = false
					room.gamemetadata.gameturnskip = false
					room.gamemetadata.gameturntime = time.Now().Add(18 * time.Second)
					room.reportuserattemptturn()
				}
				if room.gamemetadata.gameturntime.After(time.Now()) && room.gamemetadata.gameturnskip == true && room.gamemetadata.rounds > 0 { //client skipped the turn
					log.Println("Player skipped")
					room.gamemetadata.rounds = room.gamemetadata.rounds + 1
					room.gamemetadata.gameturnskip = false
					room.gamemetadata.gameturn = false
					room.gamemetadata.gameturntime = time.Now().Add(18 * time.Second)
					room.reportuserskipturn()
				}

			}
		}
	}
}

func (room *Room) registerClientinRoom(client *Client) {
	room.clients[client] = true
}

func (room *Room) unregisterClientinRoom(client *Client) {
	if _, ok := room.clients[client]; ok {
		delete(room.clients, client)
		client.Gamemetadata = &ClientGameMetadata{} //empty struct
	}
	if len(room.clients) < 2 {
		if room.gamemetadata.gamestarted == true {
			room.broadcastgamestats()
		}

	}

}

func (room *Room) broadcastToClientsInRoom(message []byte) {
	for client := range room.clients {
		client.send <- message //Client
	}
}

func (room *Room) broadcastToPreviousClientInRoomByBot(clientwholeft *Client) {
	clientlist := room.gamemetadata.clientlist
	index := 0
	var nextclient *Client
	for key, client := range clientlist {
		if client.Slug == clientwholeft.Slug {
			index = key
		}
	}
	if index+1 == len(clientlist) { //last key
		index = 0 //first client
	} else {
		index += 1 //next client
	}
	nextclient = clientlist[index]
	playername := nextclient.Name
	playerslug := nextclient.Slug
	message := &RoomBotGameMessage{
		Target:          room.name,
		Message:         fmt.Sprintf("Woops The client left who had turn so assigning you new word , <b>%v<small>@%v</small></b> <br> start with letter <b>W</b> <br>Time starts now 18 seconds ⌛", playername, playerslug),
		Action:          MessageByBot,
		WhichClientTurn: nextclient, //first one lets go
		ClientList:      clientlist,
		Letter:          "D",
		Timer:           18,
	}
	for client := range room.clients {
		client.send <- message.encode() //Client
	}
}

const welcomeMessage = "%s %s joined the room"

func (room *Room) notifyClientJoined(client *Client) {
	message := &Message{
		Action:  JoinRoomNotification,
		Target:  room.name,
		Message: fmt.Sprintf(welcomeMessage, client.Name, client.Slug),
		Sender:  client,
	}

	room.broadcastToClientsInRoom(message.encode())
}

func (room *Room) reportuserattemptturn() {
	guessedword := strings.ToLower(room.gamemetadata.wordguessed)
	letter := strings.ToLower(room.gamemetadata.letter)
	currentplayer := room.gamemetadata.whichclientturn
	nextplayer := room.getnextplayer(currentplayer)
	wordslist := room.gamemetadata.wordslist
	status := MatchWord(guessedword, wordslist, letter[0])

	resmessage := ""
	if status == "word-correct" {
		room.gamemetadata.letter = string(guessedword[len(guessedword)-1]) //correct then new letter
		resmessage = fmt.Sprintf("Bravo correct word guessed by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter)
		room.gamemetadata.whichclientturn.Gamemetadata.Score += 1
		room.gamemetadata.wordslist[guessedword] = true //add it to our word list

	} else if status == "wrong-letter" {
		resmessage = fmt.Sprintf("Word starts with wrong letter <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter)
	} else if status == "no-such-word" {
		resmessage = fmt.Sprintf("No such word exisists in our dictionary guessed by <b>%v<small>@%v</small></b> now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter)
	} else if status == "word-reused" {
		resmessage = fmt.Sprintf("This word is already guessed  <b>%v<small>@%v</small></b> so not helpfull now <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter)
	}
	room.gamemetadata.whichclientturn = nextplayer
	message := &RoomBotGameMessage{
		Target:          room.name,
		Message:         resmessage,
		Action:          MessageByBot,
		WhichClientTurn: nextplayer, //first one lets go
		ClientList:      room.gamemetadata.clientlist,
		Letter:          room.gamemetadata.letter,
		Timer:           18,
	}
	time.Sleep(1 * time.Second)
	room.broadcastToClientsInRoom(message.encode())
}

func (room *Room) reportuserskipturn() {
	currentplayer := room.gamemetadata.whichclientturn
	nextplayer := room.getnextplayer(currentplayer)
	room.gamemetadata.whichclientturn = nextplayer
	message := &RoomBotGameMessage{
		Target:          room.name,
		Message:         fmt.Sprintf("Woops The client who had turn <b>left the game</b> so assigning you new word , <b>%v<small>@%v</small></b> <br> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter),
		Action:          MessageByBot,
		WhichClientTurn: nextplayer, //first one lets go
		ClientList:      room.gamemetadata.clientlist,
		Letter:          room.gamemetadata.letter,
		Timer:           18,
	}
	time.Sleep(1 * time.Second)
	room.broadcastToClientsInRoom(message.encode())

}

func (room *Room) reportuserturnend() {
	currentplayer := room.gamemetadata.whichclientturn
	nextplayer := room.getnextplayer(currentplayer)
	room.gamemetadata.whichclientturn = nextplayer
	message := &RoomBotGameMessage{
		Target:          room.name,
		Message:         fmt.Sprintf("Word not guessed by, <b>%v<small>@%v</small></b> times up <br> now <b>%v<small>@%v</small></b> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter),
		Action:          MessageByBot,
		WhichClientTurn: nextplayer, //first one lets go
		ClientList:      room.gamemetadata.clientlist,
		Letter:          room.gamemetadata.letter,
		Timer:           18,
	}
	time.Sleep(1 * time.Second)
	room.broadcastToClientsInRoom(message.encode())

}

func (room *Room) broadcastgamestats() {
	log.Println("Game ended", room.gamemetadata.gameendtime, "-", time.Now())
	botnotify := &Message{
		Action:  SendMessageAction,
		Message: "Game ended successfully 🍾 ",
		Target:  room.name,
		Sender:  &Client{Name: "bot-of-the-room"},
	}
	room.gamemetadata.gamestarted = false
	room.gamemetadata.gameended = true
	room.broadcastToClientsInRoom(botnotify.encode())
	time.Sleep(690 * time.Millisecond)
	wordlist := []string{}
	for words := range room.gamemetadata.wordslist {
		wordlist = append(wordlist, words)
	}
	var clientlist []*Client
	for key := range room.clients {
		clientlist = append(clientlist, key)
	}
	sort.Slice(clientlist[:], func(i, j int) bool {
		return clientlist[i].Gamemetadata.Score > clientlist[j].Gamemetadata.Score
	})
	message := &RoomBotGameMessage{
		Target:     room.name,
		Message:    strings.Join(wordlist[:], ","),
		Action:     EndGameNotification,
		ClientList: clientlist,
	}
	room.broadcastToClientsInRoom(message.encode())
}
func (room *Room) getnextplayer(currentplayer *Client) *Client {
	index := 0
	clientlist := room.gamemetadata.clientlist
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
