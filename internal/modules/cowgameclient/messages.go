package cowgameclient

import (
	"encoding/json"
	"log"
)

// different messages LobbyServer can send to writepump by send channel using broadcast channel

const (
	SendMessageAction           = "send-message"
	JoinRoomAction              = "join-room"
	JoinRandomRoomAction        = "join-random-room"
	LeaveRoomAction             = "leave-room"
	ClientNameAction            = "client-name"
	StartGameAction             = "start-the-game" // Client who is RoomMaker can only make this message to readpump it will trigger StartGameNotification to all
	JoinRoomNotification        = "join-room-notify"
	ClientListNotification      = "client-list-notify" // RoomMemberCountNotification is not needed returns client list it will also have scores
	FailJoinRoomNotification    = "fail-join-room-notify"
	FoundRandomRoomNotification = "found-random-room-notify"
	KnowYourUser                = "know-yourself" //individual message witj client details send to a particulr user after he enters a room contains slug and name
	RoomMakerNotification       = "is-room-maker"
	StartGameNotification       = "room-bot-greetings" // once this message is broadcasted in room players will render in game component for us its chat component
	EndGameNotification         = "room-bot-end-game"
	SendMessageActionByBot      = "send-message-by-bot"
	MessageByBot                = "message-by-bot"
	//RoomMemberCountNotification = "clients-count-in-room-notify" //after join room or left room notify this to all clients UI will change due to this
)

type Message struct { // in readpump also can be used in qritepump
	Action  string  `json:"action"`  //action
	Message string  `json:"message"` //message
	Target  string  `json:"target"`  //target the room
	Sender  *Client `json:"sender"`  //whose readpump is used
}
type Metadata struct { //used when action is SendMessageActionByBot
	NextClientSlug string `json:"nextclientslug"`
	Status         string `json:"status"` //success,fail empty then new
	Letter         string `json:"letter"` //empty then first time
}
type ClientsinRoomMessage struct { // we are using this to return list of clients to all clients in room when register unregister happens
	Action     string    `json:"action"`     //action
	ClientList []*Client `json:"clientlist"` //message
	Target     string    `json:"target"`     //target the room
	Sender     *Client   `json:"sender"`     //whose readpump is used
}

type ClientRoomStats struct {
	Userslug  string `json:"userslug"`  //userslug this struct will be array
	Userscore int    `json:"userscore"` //userscores start with 0
}

type RoomBotGameMessage struct {
	Action          string    `json:"action"`
	Message         string    `json:"message"`
	Target          string    `json:"target"`
	ClientList      []*Client `json:"clientstats"`     //first time send by server
	WhichClientTurn *Client   `json:"whichclientturn"` //user slug whose turn is there and he can only send
	Timer           int       `json:"timer"`
	Letter          string    `json:"letter"`
}

func (message *Message) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}

func (message *ClientsinRoomMessage) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}

func (message *RoomBotGameMessage) encode() []byte {
	json, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return json
}
