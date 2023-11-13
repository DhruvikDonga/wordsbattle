package gogamemesh

import (
	"log"
	"sync"
)

const (
	MeshGlobalRoom = "mesh-global" //MeshGlobalRoom is a room where a client gets joined when he connects to a websocket
)

type MeshServerConfig struct {
	DirectBroadCast bool
}

type MeshServer interface {
	GetClients() map[string]*client

	GetRooms() map[string]*room

	GetClientsInRoom() map[string]map[string]bool

	ConnectClient(client *client)

	DisconnectClient(client *client)

	CreateRoom(name string, client string)

	DeleteRoom(name string)

	BroadcastMessage(message *Message)

	JoinClientRoom(roomname string, clientname string)

	RemoveClientRoom(roomname string, clientname string)

	//PushMessage is to push message from the code not from the UI thats broadcast
	//returns a send only channel
	PushMessage() chan<- *Message

	//ReceiveMessage is to receive message from readpumps of the clients this can be used to manipulate
	//returns a receive only channel
	RecieveMessage() <-chan *Message

	//EventTriggers Track
	//Get the updates on the clients in room changes and act accordingly
	//Returns receive only channel map[roomname]map[clientlist]bool
	EventTriggers() <-chan map[string]map[string]bool
}

// meshServer runs like workers which are light weight instead of using rooms approach this reduces weight on rooms side
// this helps for a user to connect simultaneously multiple rooms in a single go
type meshServer struct {
	gamename      string
	isbroadcaston bool
	mu            sync.RWMutex
	clients       map[string]*client
	rooms         map[string]*room
	clientsinroom map[string]map[string]bool

	clientConnect    chan *client
	clientDisconnect chan *client

	roomCreate chan []string //[clientid,roomid] who created this room to save it as a first player in that room
	roomDelete chan string

	clientJoinedRoom chan []string //[0]-->roomslug [1]-->clientslug
	clientLeftRoom   chan []string //[0]-->roomslug [1]-->clientslugs

	processMessage       chan *Message
	clientsinroomtracker chan map[string]map[string]bool

	roomdata RoomData
}

// NewMeshServer initialize new websocket server
func NewMeshServer(name string, meshconf *MeshServerConfig, rd RoomData) *meshServer {
	return &meshServer{
		gamename:      name,
		isbroadcaston: meshconf.DirectBroadCast,
		clients:       make(map[string]*client),
		rooms:         make(map[string]*room),
		clientsinroom: make(map[string]map[string]bool),

		clientConnect:    make(chan *client),
		clientDisconnect: make(chan *client),

		roomCreate: make(chan []string),
		roomDelete: make(chan string),

		clientJoinedRoom: make(chan []string),
		clientLeftRoom:   make(chan []string),

		processMessage:       make(chan *Message),                   //unbuffered channel unlike of send of client cause it will recieve only when readpump sends in it else it will block
		clientsinroomtracker: make(chan map[string]map[string]bool), //view into the maps is your room affected by client changes

		roomdata: rd,
	}
}

// Run mesh server accepting various requests
func (server *meshServer) RunMeshServer() {

	for {
		select {
		case client := <-server.clientConnect:
			server.ConnectClient(client) //add the client

		case client := <-server.clientDisconnect:
			server.DisconnectClient(client) //remove the client

		case roomcreate := <-server.roomCreate:
			server.CreateRoom(roomcreate[0], roomcreate[1]) //add the client

		case roomname := <-server.roomDelete:
			server.DeleteRoom(roomname) //remove the client

		case roomclient := <-server.clientJoinedRoom:
			server.JoinClientRoom(roomclient[0], roomclient[1]) //add the client to room

		case roomclient := <-server.clientLeftRoom:
			server.RemoveClientRoom(roomclient[0], roomclient[1]) //remove the client from room

		case message := <-server.processMessage: //this broadcaster will broadcast to all clients
			log.Println("Websocket broadcast", message)
			if server.isbroadcaston {
				server.BroadcastMessage(message) //broadcast the message from readpump
			}

		}
	}
}

func (server *meshServer) GetClients() map[string]*client {
	return server.clients
}

func (server *meshServer) GetRooms() map[string]*room {
	return server.rooms
}

func (server *meshServer) GetClientsInRoom() map[string]map[string]bool {
	return server.clientsinroom
}

func (server *meshServer) PushMessage() chan<- *Message {
	return server.processMessage
}

func (server *meshServer) RecieveMessage() <-chan *Message {
	return server.processMessage
}

func (server *meshServer) EventTriggers() <-chan map[string]map[string]bool {
	return server.clientsinroomtracker
}

func (server *meshServer) ConnectClient(client *client) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.clients[client.slug] = client
	server.JoinClientRoom(MeshGlobalRoom, client.slug) //join this default to a room this is a global room kind of main lobby
}

func (server *meshServer) DisconnectClient(client *client) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for roomname, clientsmap := range server.clientsinroom {
		delete(clientsmap, client.slug)
		server.clientsinroomtracker <- server.clientsinroom
		if len(clientsmap) == 0 && roomname != MeshGlobalRoom {
			delete(server.clientsinroom, roomname)
		}
	}

	delete(server.clients, client.slug)

}

func (server *meshServer) CreateRoom(name string, client string) {

	room := NewRoom(name, client, server.roomdata, server)
	server.mu.Lock()
	defer server.mu.Unlock()
	server.rooms[room.slug] = room //add it to server list of rooms

}

func (server *meshServer) DeleteRoom(name string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	if r, ok := server.rooms[name]; ok {
		close(r.stopped)
		delete(server.rooms, name)
	}

}

func (server *meshServer) BroadcastMessage(message *Message) {
	server.mu.RLock()
	defer server.mu.RUnlock()
	jsonBytes := message.encode()
	if message.IsTargetClient {
		client := server.clients[message.Target]

		client.send <- jsonBytes
	} else {
		clients := server.clientsinroom[message.Target]

		for c := range clients {
			client := server.clients[c]
			client.send <- jsonBytes
		}
	}

}

func (server *meshServer) JoinClientRoom(roomname string, clientname string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for roomkey := range server.clientsinroom {
		if roomkey == roomname {
			server.clientsinroom[roomkey][clientname] = true
			server.clientsinroomtracker <- server.clientsinroom
			break
		}
	}
}

func (server *meshServer) RemoveClientRoom(roomname string, clientname string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for roomkey, clientsmap := range server.clientsinroom {
		if roomkey == roomname {
			delete(clientsmap, clientname)
			server.clientsinroomtracker <- server.clientsinroom
			if len(clientsmap) == 0 && roomname != MeshGlobalRoom {
				delete(server.clientsinroom, roomname)
			}
			break
		}
	}
}
