package gomeshstream

import (
	"log"
	"sync"
)

type RoomData interface {
	//HandleRoomData use your struct which has all the data related to room and do the changes accordingly
	HandleRoomData(room Room, server MeshServer)
}

type Room interface {
	GetRoomSlugInfo() string

	GetRoomMakerInfo() string

	RoomStopped() <-chan struct{}

	ConsumeRoomMessage() <-chan *Message

	EventTriggers() <-chan []string

	BroadcastMessage(message *Message)
}

type room struct {
	mu sync.RWMutex

	id                int
	server            *meshServer
	slug              string
	createdby         string //client id who created it
	stopped           chan struct{}
	roomdata          RoomData
	consumeMessage    chan *Message
	clientInRoomEvent chan []string
	clientsinroom     map[string]*client
}

func NewRoom(roomslug string, clientslug string, rd RoomData, srv *meshServer) *room {
	srv.roomcnt += 1

	r := &room{
		mu:                sync.RWMutex{},
		id:                srv.roomcnt,
		slug:              roomslug,
		createdby:         clientslug,
		stopped:           make(chan struct{}, 1),
		roomdata:          rd,
		server:            srv,
		consumeMessage:    make(chan *Message, 1),
		clientInRoomEvent: make(chan []string, 1),
		clientsinroom:     make(map[string]*client),
	}
	go func() {
		r.roomdata.HandleRoomData(r, srv)
	}()
	log.Println("room created and running", roomslug)

	return r
}

func (room *room) GetRoomSlugInfo() string {
	return room.slug
}

func (room *room) GetRoomMakerInfo() string {
	return room.createdby
}

func (room *room) RoomStopped() <-chan struct{} {
	return room.stopped
}

func (room *room) ConsumeRoomMessage() <-chan *Message {
	return room.consumeMessage
}

func (room *room) EventTriggers() <-chan []string {
	return room.clientInRoomEvent
}

func (room *room) BroadcastMessage(message *Message) {
	room.mu.RLock()
	defer room.mu.RUnlock()
	jsonBytes := message.Encode()
	log.Println("Broadcasting message from room ----", string(jsonBytes))
	if message.IsTargetClient {

		client := room.clientsinroom[message.Target]
		log.Println("Pushing to client :-", client.slug)

		client.send <- jsonBytes
	} else {
		clients := room.clientsinroom
		log.Println("Pushing to clients :-", clients)
		for _, c := range clients {
			c.send <- jsonBytes
		}
	}

}
