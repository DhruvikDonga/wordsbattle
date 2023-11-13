package gogamemesh

type RoomData interface {
	//HandleRoomData use your struct which has all the data related to room and do the changes accordingly
	HandleRoomData(room Room, server MeshServer)
}

type Room interface {
	GetRoomSlugInfo() string

	RoomStopped() <-chan struct{}
}

type room struct {
	server    *meshServer
	slug      string
	createdby string //client id who created it
	stopped   chan struct{}
	roomdata  RoomData
}

func NewRoom(roomslug string, clientslug string, rd RoomData, srv *meshServer) *room {

	r := &room{
		slug:      roomslug,
		createdby: clientslug,
		stopped:   make(chan struct{}),
		roomdata:  rd,
		server:    srv,
	}

	go func() {
		r.roomdata.HandleRoomData(r, srv)
	}()
	return r
}

func (room *room) GetRoomSlugInfo() string {
	return room.slug
}

func (room *room) RoomStopped() <-chan struct{} {
	return room.stopped
}
