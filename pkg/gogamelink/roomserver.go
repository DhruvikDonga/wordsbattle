package gogamelink

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// room server is a goroutine created by the chatServer when "join-room" message is catched by read-pump it either joins the user to exisiting room or create a new room
// room will handle new rooms associated clients register, unregister a client into it and also broadcast to clients

type GameRoomMetadata struct { //this can change as per game and it should be initialized when game is started
	gamestarted     bool
	gameendtime     time.Time
	gameended       bool
	gamechatevent   bool      //this is trial purpose if this event is false will have to do it
	gamechattime    time.Time //this is chat time in cases
	gameturn        bool      //if set to true user has played his turn
	gameturnskip    bool      //if set to true user skipped his turn might be he left the server
	gameturntime    time.Time //this is recursive will set to null then start again can be different in cases
	clientlist      []*client
	whichclientturn *client
	rounds          int
}

type RoomStateHandler interface {
	//SendMessageToRoom function is spawned in a goroutine that act as a service in a room which can be used to handle the client responses process the algorithm and then forward it . You can use it as a bot too.
	SendMessageToRoom(room Room)
}

type Room interface {
	//BroadcastToClientsInRoom send message to all the clients present in a room
	BroadcastToClientsInRoom(message []byte)

	//BroadcastToAClientInRoom sends message to a particular clients in a room
	// clientid : message
	BroadcastToAClientInRoom(message map[string][]byte)

	//Room is stopped
	RoomIsStopped() <-chan bool

	//Channel forwarder nothing more nothing less
	MessageRecievedByBot() <-chan []byte

	//GetRoomID returns ID of the room can be usefull if you have your map to store RoomStateData cause package does not handles it
	GetRoomID() string
}

type room struct {
	sync.RWMutex
	name                    string
	clients                 map[string]bool //slugid string
	wsServer                *LobbyServer    //keep reference of webserver to every client
	register                chan *client
	unregister              chan *client
	broadcast               chan []byte            //message send in a room
	broadcastToClientinRoom chan map[string][]byte //message to a client map[clientslug]json[]byte
	broadcastToBot          chan []byte
	stoproom                chan bool //end the gameserver routine
	randomgame              bool      //true then it is random game 2 players only once player connect then start the game automatic no lobby system
	roommaker               *client   //room maker will get a seperate message which will make him capable to start the game
	gamemetadata            *GameRoomMetadata
	playerlimit             int
	stateHandler            RoomStateHandler
}

func NewRoom(name string, gameServer *LobbyServer, roommaker *client, roomStateHandle RoomStateHandler) *room {
	r := &room{
		name:                    name,
		wsServer:                gameServer,
		clients:                 make(map[string]bool),
		register:                make(chan *client),
		unregister:              make(chan *client),
		broadcast:               make(chan []byte), //unbuffered channel unlike of send of client cause it will recieve only when readpump sends in it else it will block
		broadcastToClientinRoom: make(chan map[string][]byte),
		broadcastToBot:          make(chan []byte, 1),
		stoproom:                make(chan bool, 1),
		//randomgame:               true, //set it during create room
		roommaker:    roommaker,
		stateHandler: roomStateHandle,
		gamemetadata: &GameRoomMetadata{gamestarted: false, gamechatevent: false, gameturn: false, gameturnskip: false, rounds: 0}, //game not started in room
	}
	go r.runRoom()
	go r.stateHandler.SendMessageToRoom(r)
	return r
}

// Run websocket server accepting various requests
func (r *room) runRoom() {
	//gameStateTicker := time.NewTicker(5 * time.Millisecond) //need update for room every 5 millisecond
	for {
		select {
		case c := <-r.register:
			r.registerClientinRoom(c.Slug) //add the client
			r.notifyClientJoined(c.Slug)   // notify to clients in room
			if r.randomgame {              //notify the roomname to client
				message := &Message{
					Action:  FoundRandomRoomNotification,
					Message: fmt.Sprintf(r.name),
				}
				c.send <- message.encode()
			}

		case c := <-r.unregister:
			r.unregisterClientinRoom(c.Slug) //remove the client
			//clients in room notification
			var clientlist []*client
			for key := range r.clients {
				clientlist = append(clientlist, r.wsServer.clients[key])
			}
			message := &ClientsinRoomMessage{
				Action:     ClientListNotification,
				Target:     r.name,
				ClientList: clientlist,
			}
			r.BroadcastToClientsInRoom(message.encode())
			c.ingame = false // set it is not in any game
			c.room = &room{} // set empty room
			log.Println("Check if room is to be deleted")
			if len(r.clients) == 0 {
				log.Println("room shutdown", r.name)
				r.wsServer.DeleteRoom(r)
				return
			}
		case message := <-r.broadcast:
			r.BroadcastToClientsInRoom(message) //broadcast the message from readpump to a room clients only
		case message := <-r.broadcastToClientinRoom: //send message to certain clients in room
			r.BroadcastToAClientInRoom(message)
		case <-r.broadcastToBot:
			r.MessageRecievedByBot()
		case <-r.stoproom:
			log.Println("room shutdown", r.name)
			r.wsServer.DeleteRoom(r)
			return
		}
	}

}

func (room *room) registerClientinRoom(clientslug string) {
	room.Lock()
	defer room.Unlock()
	room.clients[clientslug] = true
}

func (room *room) unregisterClientinRoom(clientslug string) {
	room.Lock()
	defer room.Unlock()
	delete(room.clients, clientslug)

}

func (r *room) GetRoomID() string {
	return r.name
}

// RoomIsStopped returns true use it in select statement for SendMessageToRoom RoomStateHandler
func (r *room) RoomIsStopped() <-chan bool {
	return r.stoproom
}

func (r *room) MessageRecievedByBot() <-chan []byte {
	return r.broadcastToBot
}

func (room *room) BroadcastToClientsInRoom(message []byte) {
	room.RLock()
	defer room.RUnlock()
	for clientslug := range room.clients {
		room.wsServer.RLock()
		room.wsServer.clients[clientslug].send <- message //Client
		room.wsServer.RUnlock()
	}
}
func (room *room) BroadcastToAClientInRoom(message map[string][]byte) { //this message should be containing one key and one value though
	for clientid, message := range message {
		room.wsServer.RLock()
		client := room.wsServer.clients[clientid]
		room.wsServer.RUnlock()
		client.send <- message //Client
	}
}

const welcomeMessage = "%s joined the room"

func (room *room) notifyClientJoined(clientslug string) {
	client := room.wsServer.clients[clientslug]
	message := &Message{
		Action:  JoinRoomNotification,
		Target:  room.name,
		Message: fmt.Sprintf(welcomeMessage, client.Slug),
		Sender:  client,
	}

	room.BroadcastToClientsInRoom(message.encode())
}
