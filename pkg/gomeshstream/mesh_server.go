package gomeshstream

import (
	"log"
	"sync"
)

const (
	MeshGlobalRoom = "mesh-global" //MeshGlobalRoom is a room where a client gets joined when he connects to a websocket . MeshGlobalRoom facilitates creation of rooms
)

type MeshServerConfig struct {
	DirectBroadCast bool
}

type MeshServer interface {
	GetClients() map[string]*client

	GetRooms() []string

	GetGameName() string

	GetClientsInRoom() map[string]map[string]*client

	ConnectClient(client *client)

	DisconnectClient(client *client)

	CreateRoom(name string, client string, rd RoomData)

	DeleteRoom(name string)

	//BroadcastMessage(message *Message)

	JoinClientRoom(roomname string, clientname string, rd RoomData)

	RemoveClientRoom(roomname string, clientname string)

	//PushMessage is to push message from the code not from the UI thats broadcast
	//returns a send only channel
	PushMessage() chan<- *Message

	//ReceiveMessage is to receive message from readpumps of the clients this can be used to manipulate
	//returns a receive only channel
	RecieveMessage() <-chan *Message

	//EventTriggers Track
	//Get the updates on the clients in room changes and act accordingly
	//Returns receive only channel []string length of 3 [0]-->event type [1]-->roomname [1]-->clientslug
	//event types :- client-joined-room , client-left-room
	EventTriggers() <-chan []string

	JoinARoom() chan<- []interface{}
}

// meshServer runs like workers which are light weight instead of using rooms approach this reduces weight on rooms side
// this helps for a user to connect simultaneously multiple rooms in a single go
type meshServer struct {
	mu sync.RWMutex

	gamename         string
	isbroadcaston    bool
	clients          map[string]*client
	rooms            map[string]*room
	clientsinroom    map[string]map[string]*client
	roomcnt          int
	clientConnect    chan *client
	clientDisconnect chan *client

	roomCreate chan []string //[clientid,roomid] who created this room to save it as a first player in that room
	roomDelete chan string

	clientJoinedRoom chan []interface{} //[0]-->roomslug [1]-->clientslug [2]--> RoomData
	clientLeftRoom   chan []string      //[0]-->roomslug [1]-->clientslugs

	processMessage    chan *Message
	clientInRoomEvent chan []string //[event type,room name, client slug] , client-joined-room, client-left-room

	roomdata RoomData
}

// NewMeshServer initialize new websocket server
func NewMeshServer(name string, meshconf *MeshServerConfig, rd RoomData) *meshServer {
	server := &meshServer{
		mu:            sync.RWMutex{},
		gamename:      name,
		roomcnt:       0,
		isbroadcaston: meshconf.DirectBroadCast,
		clients:       make(map[string]*client),
		rooms:         make(map[string]*room),
		clientsinroom: make(map[string]map[string]*client),

		clientConnect:    make(chan *client, 1),
		clientDisconnect: make(chan *client, 1),

		roomCreate: make(chan []string, 1),
		roomDelete: make(chan string, 1),

		clientJoinedRoom: make(chan []interface{}, 1),
		clientLeftRoom:   make(chan []string, 1),

		processMessage:    make(chan *Message, 1), //unbuffered channel unlike of send of client cause it will recieve only when readpump sends in it else it will block
		clientInRoomEvent: make(chan []string, 1), //view into the maps is your room affected by client changes

		roomdata: rd,
	}
	r := &room{
		id:                server.roomcnt,
		slug:              MeshGlobalRoom,
		createdby:         "Gawd",
		stopped:           make(chan struct{}),
		roomdata:          rd,
		server:            server,
		consumeMessage:    make(chan *Message, 1),
		clientInRoomEvent: make(chan []string, 1),
	}
	server.roomcnt += 1

	server.rooms[MeshGlobalRoom] = r
	go func() {
		server.roomdata.HandleRoomData(r, server)
	}()
	go func() {
		server.RunMeshServer()
	}()

	return server
}

// Run mesh server accepting various requests
func (server *meshServer) RunMeshServer() {
	log.Println("Mesh Server ", server.gamename, " started")
	for {
		select {
		case client := <-server.clientConnect:
			server.ConnectClient(client) //add the client

		case client := <-server.clientDisconnect:
			server.DisconnectClient(client) //remove the client

		case roomcreate := <-server.roomCreate:
			server.CreateRoom(roomcreate[0], roomcreate[1], server.roomdata) //add the client

		case roomname := <-server.roomDelete:
			server.DeleteRoom(roomname) //remove the client

		case roomclient := <-server.clientJoinedRoom:
			log.Println("Request from a client to join a room")
			server.JoinClientRoom(roomclient[0].(string), roomclient[1].(string), roomclient[2].(RoomData)) //add the client to room

		case roomclient := <-server.clientLeftRoom:
			server.RemoveClientRoom(roomclient[0], roomclient[1]) //remove the client from room

		case message := <-server.processMessage: //this broadcaster will broadcast to all clients
			//log.Println("Websocket broadcast", message)
			// if server.isbroadcaston {
			// 	server.BroadcastMessage(message) //broadcast the message from readpump
			// }
			roomtosend := server.rooms[message.Target]
			select {
			case roomtosend.consumeMessage <- message:
				//log.Println("Room name ", roomtosend.slug, "Created by ", roomtosend.createdby)
			default:
				log.Println("Failed to send  to Room name ", roomtosend.slug)

			}
		}
	}
}

func (server *meshServer) GetClients() map[string]*client {
	return server.clients
}

func (server *meshServer) GetGameName() string {
	return server.gamename
}

func (server *meshServer) GetRooms() []string {
	log.Println("Getting the rooms list")
	roomslist := []string{}

	server.mu.Lock()
	defer server.mu.Unlock()
	for room := range server.rooms {
		roomslist = append(roomslist, room)
	}
	return roomslist
}

func (server *meshServer) GetClientsInRoom() map[string]map[string]*client {
	server.mu.Lock()
	res := server.clientsinroom
	server.mu.Unlock()
	return res
}

func (server *meshServer) PushMessage() chan<- *Message {
	return server.processMessage
}

func (server *meshServer) JoinARoom() chan<- []interface{} {
	return server.clientJoinedRoom
}

func (server *meshServer) RecieveMessage() <-chan *Message {
	return server.processMessage
}

func (server *meshServer) EventTriggers() <-chan []string {
	return server.clientInRoomEvent
}

func (server *meshServer) ConnectClient(client *client) {
	server.mu.Lock()
	server.clients[client.slug] = client
	server.mu.Unlock()
	server.JoinClientRoom(MeshGlobalRoom, client.slug, server.roomdata) //join this default to a room this is a global room kind of main lobby
}

func (server *meshServer) DisconnectClient(client *client) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for roomname, clientsmap := range server.clientsinroom {
		if _, ok := clientsmap[client.slug]; ok {
			delete(clientsmap, client.slug)
			delete(server.rooms[roomname].clientsinroom, client.slug)
			log.Println("removing client from the room")
			if roomname != MeshGlobalRoom {
				select {
				case server.rooms[roomname].clientInRoomEvent <- []string{"client-left-room", roomname, client.slug}:
					log.Println("client ", client.slug, " left a room ", roomname)
				default:
					log.Println("Failed to trigger left room trigger for client ", client.slug, " in room", roomname)
				}
				if len(clientsmap) == 0 && roomname != MeshGlobalRoom {
					delete(server.clientsinroom, roomname)
					//server.DeleteRoom(roomname)
					if r, ok := server.rooms[roomname]; ok {
						log.Println("Closing room")
						close(r.stopped)
						delete(server.rooms, roomname)
					}
				}
			}
		}
	}

	delete(server.clients, client.slug)

}

func (server *meshServer) CreateRoom(name string, client string, rd RoomData) {

	room := NewRoom(name, client, rd, server)

	server.mu.Lock()
	server.rooms[room.slug] = room //add it to server list of rooms
	server.mu.Unlock()

}

func (server *meshServer) DeleteRoom(name string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	if r, ok := server.rooms[name]; ok {
		close(r.stopped)
		delete(server.rooms, name)
	}

}

// func (server *meshServer) BroadcastMessage(message *Message) {
// 	server.mu.RLock()
// 	defer server.mu.RUnlock()
// 	jsonBytes := message.Encode()
// 	log.Println("Broadcasting message ----", string(jsonBytes))
// 	if message.IsTargetClient {

// 		client := server.clients[message.Target]
// 		log.Println("Pushing to client :-", client.slug)

// 		client.send <- jsonBytes
// 	} else {
// 		clients := server.clientsinroom[message.Target]
// 		log.Println("Pushing to clients :-", clients)
// 		for c := range clients {
// 			client := server.clients[c]
// 			client.send <- jsonBytes
// 		}
// 	}

// }

func (server *meshServer) JoinClientRoom(roomname string, clientname string, rd RoomData) {
	noroom := false
	server.mu.RLock()
	if _, ok := server.rooms[roomname]; !ok {
		noroom = true

	}
	server.mu.RUnlock()
	if noroom {
		log.Printf("Create room Details: %+v\n ", rd)
		server.CreateRoom(roomname, clientname, rd)
	}
	server.mu.Lock()
	for roomkey := range server.rooms {
		if roomkey == roomname {
			if clientinroom, ok := server.clientsinroom[roomkey]; ok {
				clientinroom[clientname] = server.clients[clientname]
			} else {
				server.clientsinroom[roomkey] = map[string]*client{}
				server.clientsinroom[roomkey][clientname] = server.clients[clientname]
			}
			//copy it to the room and keep it updated
			server.rooms[roomname].clientsinroom = server.clientsinroom[roomname]
			select {
			case server.rooms[roomname].clientInRoomEvent <- []string{"client-joined-room", roomname, clientname}:
				log.Println("client ", clientname, " joined a room ", roomname)
			default:
				log.Println("Failed to trigger join room trigger for client ", clientname, " in room", roomname)
			}

			server.mu.Unlock()
			return

		}
	}

}

func (server *meshServer) RemoveClientRoom(roomname string, clientname string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	if clientsmap, ok := server.clientsinroom[roomname]; ok {
		delete(clientsmap, clientname)
		delete(server.rooms[roomname].clientsinroom, clientname)
		server.clientInRoomEvent <- []string{"client-left-room", roomname, clientname}
		select {
		case server.rooms[roomname].clientInRoomEvent <- []string{"client-left-room", roomname, clientname}:
			log.Println("client ", clientname, " removed from a room ", roomname)
		default:
			log.Println("Failed to trigger left room trigger for client ", clientname, " in room", roomname)
		}
		if len(clientsmap) == 0 && roomname != MeshGlobalRoom {
			delete(server.clientsinroom, roomname)
			//server.DeleteRoom(roomname)
			if r, ok := server.rooms[roomname]; ok {
				log.Println("Closing room")
				close(r.stopped)
				delete(server.rooms, roomname)
			}
		}
	}

}
