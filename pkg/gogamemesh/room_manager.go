package gogamemesh

import "log"

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
}

type room struct {
	id                int
	server            *meshServer
	slug              string
	createdby         string //client id who created it
	stopped           chan struct{}
	roomdata          RoomData
	consumeMessage    chan *Message
	clientInRoomEvent chan []string
}

func NewRoom(roomslug string, clientslug string, rd RoomData, srv *meshServer) *room {
	srv.roomcnt += 1

	r := &room{
		id:                srv.roomcnt,
		slug:              roomslug,
		createdby:         clientslug,
		stopped:           make(chan struct{}, 1),
		roomdata:          rd,
		server:            srv,
		consumeMessage:    make(chan *Message, 1),
		clientInRoomEvent: make(chan []string, 1),
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
