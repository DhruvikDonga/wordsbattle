//uses gogame link package

package cowgameclient

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/DhruvikDonga/wordsbattle/pkg/gogamelink"

	"log"
)

type ClientScores struct {
	ClientID string
	Score    uint64
}

type CoWGameRoomState struct {
	gamestarted     bool
	gameendtime     time.Time
	gameended       bool
	gamechatevent   bool            //this is trial purpose if this event is false will have to do it
	gamechattime    time.Time       //this is chat time in cases
	gameturn        bool            //if set to true user has played his turn
	gameturnskip    bool            //if set to true user skipped his turn might be he left the server
	gameturntime    time.Time       //this is recursive will set to null then start again can be different in cases
	wordslist       map[string]bool //word and meaning
	clientlist      []*ClientScores
	whichclientturn string
	letter          string //[0]
	wordguessed     string
	rounds          int
}

// Stores all Room types by their name.
var RoomStateManager = struct {
	sync.Mutex
	RoomStates map[string]*CoWGameRoomState
}{
	RoomStates: make(map[string]*CoWGameRoomState, 0),
}

type ClientCustomMessage struct {
	Action  string  `json:"action"`  //action
	Message string  `json:"message"` //message
	Target  string  `json:"target"`  //target the room
	Sender  *Client `json:"sender"`  //whose readpump is used
}

type RoomGameBot struct {
	Action string `json:"action"` //action
}

// HandleRoomMessages
func (custommessage ClientCustomMessage) HandleMessage(c gogamelink.Client, jsonMessage []byte, clientid string, room string) {
	if err := json.Unmarshal(jsonMessage, &custommessage); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
	}

	switch custommessage.Action {
	case JoinRoomAction:
		roomname := custommessage.Message //this message string will have roomname string
		c.AddClientToRoom(roomname)

	case JoinRandomRoomAction:
		roomname := custommessage.Message
		c.HandleJoinRandomRoomMessage(roomname)

	case LeaveRoomAction:
		c.HandleLeaveRoomMessage()

	case StartGameAction: //its case of friends where one can start the game generally UI only enables this option to a particular case and sends a message of that type we are taking that message and brodcasting in room
		roomname := custommessage.Target
		cmessage := &ClientCustomMessage{ //imp will change ui
			Action: StartGameNotification,
			Target: roomname,
		}
		c.HandleSendMessageToRoom(roomname, cmessage.encode())

	case "received-a-word":
		roomname := custommessage.Target
		cmessage := &ClientCustomMessage{ //imp will change ui
			Action:  StartGameNotification,
			Target:  roomname,
			Message: custommessage.Message,
		}
		c.HandleSendMessageToBotOfRoom(roomname, cmessage.encode())
	}

}

func (message *ClientCustomMessage) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}

func (roomHandle RoomGameBot) SendMessageToRoom(r gogamelink.Room) { //This will send message to room in 3 minutes
	endTime := time.After(3 * time.Minute)
	turnSwitchAfter := time.After(10 * time.Second)
	greetingsMessageSend := false
	//perRoundTicker := 10 //tick tick check for 10 seconds 10 second a round
	for {
		select {

		case message := <-r.MessageRecievedByBot():
			//algorthm check and return response restart the ticker
			log.Println(string(message))

			//ones send a next rounder start a ticker
			turnSwitchAfter = time.After(10 * time.Second)

		case <-endTime:
			message := &RoomGameBot{
				Action: EndGameNotification,
			}
			json, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
			}
			r.BroadcastToClientsInRoom(json)

		case <-turnSwitchAfter: //if this comes up it means time switcher ends and restart the clock
			//it must have failed to get the move on this one so notify
			// message := &RoomBotGameMessage{
			// 	Message:         fmt.Sprintf("Word not guessed by, <b>%v<small>@%v</small></b> times up <br> now <b>%v<small>@%v</small></b> start with letter <b>%v</b> <br>Time starts now 18 seconds ⌛", currentplayer.Name, currentplayer.Slug, nextplayer.Name, nextplayer.Slug, room.gamemetadata.letter),
			// 	Action:          MessageByBot,
			// }
			failedtheturn := &RoomBotGameMessage{
				Action:  SendMessageAction,
				Target:  r.GetRoomID(),
				Message: "Turn switched <b>Bot<small>@room-<small>" + r.GetRoomID() + "</small></small></b> here . I will be having my 👀 eyes over you if you playing fair, update the score board and assign you new letter",
				Timer:   18,
			}
			r.BroadcastToClientsInRoom(failedtheturn.encode())
			turnSwitchAfter = time.After(10 * time.Second)

		case <-r.RoomIsStopped(): //room is stopped so close this routine
			return

		default: //use it only to send messages through this bot
			if !greetingsMessageSend {
				botgreetingsmessage := &Message{
					Action:  SendMessageAction,
					Target:  r.GetRoomID(),
					Sender:  &Client{Name: "bot-of-the-room"},
					Message: "Yo this is <b>Bot<small>@room-<small>" + r.GetRoomID() + "</small></small></b> here . I will be having my 👀 eyes over you if you playing fair, update the score board and assign you new letter",
				}
				r.BroadcastToClientsInRoom(botgreetingsmessage.encode())
			}
		}
	}

}
