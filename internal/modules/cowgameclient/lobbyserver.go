// lobbyserver make sure we can send and receive messages with our connected client. In order to keep track of the connected clients at the server.
package cowgameclient

import (
	"encoding/json"
	"log"
	"strconv"
)

type LobbyServer struct {
	clients           map[*Client]bool          // clients registered in server wsclients.go
	register          chan *Client              // request to register a client
	unregister        chan *Client              // request to unregister a client
	broadcast         chan []byte               // broadcast channel listens for messages, sent by the client readPump. It in turn pushes this messages in the send channel of all the clients registered.
	broadcastToClient chan map[*Client]*Message // broadcast to only particular client
	rooms             map[*Room]bool            // rooms created
}

// NewLobbyServer initialize new websocket server
func NewLobbyServer() *LobbyServer {
	return &LobbyServer{
		clients:           make(map[*Client]bool),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan []byte), //unbuffered channel unlike of send of client cause it will recieve only when readpump sends in it else it will block
		broadcastToClient: make(chan map[*Client]*Message),
		rooms:             make(map[*Room]bool),
	}
}

// Run lobby server accepting various requests
func (server *LobbyServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client) //add the client
			server.broadcastActiveMessage()
		case client := <-server.unregister:
			server.unregisterClient(client) //remove the client
			server.broadcastActiveMessage()

		case message := <-server.broadcast: //this broadcaster will broadcast to all clients
			log.Println("Websocket broadcast", message)
			server.broadcastToClients(message) //broadcast the message from readpump

		case message := <-server.broadcastToClient: //this broadcaster will broadcast to particular clients
			log.Println("Websocket broadcast", message)
			server.broadcastToAClient(message)
		}
	}
}

func (server *LobbyServer) registerClient(client *Client) {

	server.clients[client] = true
}

func (server *LobbyServer) unregisterClient(client *Client) {
	if _, ok := server.clients[client]; ok {
		delete(server.clients, client)
	}

}

func (server *LobbyServer) broadcastToClients(message []byte) {
	for client := range server.clients {
		client.send <- message //Client
	}
}
func (server *LobbyServer) broadcastToAClient(message map[*Client]*Message) { //this message should be containing one key and one value though
	for client, message := range message {
		client.send <- message.encode() //Client
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
	server.broadcastToClients(jsonStr)
}

func (server *LobbyServer) findRoom(name string) *Room {
	var foundroom *Room
	for room := range server.rooms {
		if room.name == name {
			foundroom = room
			log.Println("found room", len(room.clients))

			break
		}
	}
	return foundroom
}

func (server *LobbyServer) createRoom(name string, client *Client, playerlimit int, israndom bool) *Room {

	room := NewRoom(name, server, client)
	room.playerlimit = playerlimit //play with friend
	room.randomgame = israndom
	go room.RunRoom()         //run this room in a routine
	server.rooms[room] = true //add it to server list of rooms
	log.Printf("started goroutine for room %v by %v\n", room.name, room.roommaker.Slug)
	return room

}
func (server *LobbyServer) deleteRoom(room *Room) {
	log.Println("room deleted", room.name)
	if _, ok := server.rooms[room]; ok {
		delete(server.rooms, room)
	}

}
