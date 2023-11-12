// lobbyserver make sure we can send and receive messages with our connected client. In order to keep track of the connected clients at the server.
package gogamelink

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
)

type LobbyServer struct {
	sync.RWMutex
	name              string
	clients           map[string]*client      // clients registered in server wsclients.go map[ID]Client
	register          chan *client            // request to register a client
	unregister        chan *client            // request to unregister a client
	broadcast         chan []byte             // broadcast channel listens for messages, sent by the client readPump. It in turn pushes this messages in the send channel of all the clients registered.
	broadcastToClient chan map[*client][]byte // broadcast to only particular client
	rooms             map[string]*room        // rooms created
	roomStateHandler  RoomStateHandler        //mulitple routines to produce messages on a room using tickers for example
}

// RunNewLobbyServer initialize new websocket server
func RunNewLobbyServer(name string, roomState RoomStateHandler) *LobbyServer {
	l := &LobbyServer{
		name:              name,
		clients:           make(map[string]*client),
		register:          make(chan *client),
		unregister:        make(chan *client),
		broadcast:         make(chan []byte), //unbuffered channel unlike of send of client cause it will recieve only when readpump sends in it else it will block
		broadcastToClient: make(chan map[*client][]byte),
		rooms:             make(map[string]*room),
		roomStateHandler:  roomState,
	}
	go l.Run()
	return l
}

// Run lobby server accepting various requests
func (server *LobbyServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.RegisterClient(client) //add the client
			server.broadcastActiveMessage()
		case client := <-server.unregister:
			server.UnregisterClient(client) //remove the client
			server.broadcastActiveMessage()

		case message := <-server.broadcast: //this broadcaster will broadcast to all clients
			log.Println("Websocket broadcast", message)
			server.BroadcastToClients(message) //broadcast the message from readpump

		case message := <-server.broadcastToClient: //this broadcaster will broadcast to particular clients
			log.Println("Websocket broadcast", message)
			server.BroadcastToAClient(message)
		}
	}
}

func (server *LobbyServer) RegisterClient(client *client) {

	server.clients[client.Slug] = client
}

func (server *LobbyServer) UnregisterClient(client *client) {
	delete(server.clients, client.Slug)
}

func (server *LobbyServer) BroadcastToClients(message []byte) {
	for _, client := range server.clients {
		client.send <- message //Client
	}
}
func (server *LobbyServer) BroadcastToAClient(message map[*client][]byte) { //this message should be containing one key and one value though
	for client, message := range message {
		client.send <- message //Client
	}
}

func (server *LobbyServer) broadcastActiveMessage() {
	activeusers := map[string]string{
		"activenow": strconv.Itoa(len(server.clients)),
	}
	jsonStr, err := json.Marshal(activeusers)
	if err != nil {
		log.Printf("Error: %s", err.Error())
	} else {
		log.Println(string(jsonStr))
	}
	server.BroadcastToClients(jsonStr)
}

func (server *LobbyServer) FindRoom(name string) *room {
	server.RLock()
	defer server.RUnlock()
	var foundroom *room
	for roomname, room := range server.rooms {
		if roomname == name {
			foundroom = room
			log.Println("found room", len(room.clients))

			break
		}
	}
	return foundroom
}

func (server *LobbyServer) CreateRoom(name string, client *client, playerlimit int, israndom bool, roomStateHandler RoomStateHandler) *room {
	server.Lock()
	defer server.Unlock()
	room := NewRoom(name, server, client, roomStateHandler)
	room.playerlimit = playerlimit //play with friend
	room.randomgame = israndom
	go room.runRoom()              //run this room in a routine
	server.rooms[room.name] = room //add it to server list of rooms
	log.Printf("started goroutine for room %v by %v\n", room.name, room.roommaker.Slug)
	return room

}
func (server *LobbyServer) DeleteRoom(r *room) {
	server.Lock()
	defer server.Unlock()
	log.Println("room deleted", r.name)
	delete(server.rooms, r.name)

}
