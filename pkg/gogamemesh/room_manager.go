package gogamemesh

type RoomWatcher interface {
	WatchRoom(room Room)

	CloseWatcher()
}

type Room interface {
}

type room struct {
	server    *meshServer
	slug      string
	createdby string //client id who created it
	watch     RoomWatcher
}

func NewRoom(roomslug string, clientslug string, s *meshServer) *room {

	r := &room{
		server:    s,
		slug:      roomslug,
		createdby: clientslug,
	}
	go r.watch.WatchRoom(r)
	return r
}
