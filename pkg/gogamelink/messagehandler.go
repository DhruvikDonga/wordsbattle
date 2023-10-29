package gogamelink

import (
	"fmt"
	"log"
	"strings"
)

// handleJoinRoomMessage will handle joining a room.register <- Client
func (c *client) HandleJoinRoomMessage(roomname string) {
	room := c.wsServer.findRoom(roomname) //this will be usefull is player is playing with a friend
	// room := client.wsServe.findidealroom() //go through list of ideal rooms  first then create a new random room
	if room == nil {
		f := func(r rune) bool {
			return (r < 'A' || r > 'z') && (r < '0' || r > '9')
		}
		if len(roomname) != 10 || strings.IndexFunc(roomname, f) != -1 {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Target:  room.name,
				Message: fmt.Sprintf("Room name %v is not valid ", room.name),
			}
			msg := message.encode()
			c.wsServer.broadcastToClient <- map[*client][]byte{c: msg}
			return
		}
		room = c.wsServer.createRoom(roomname, c, 10, false, c.wsServer.roomStateHandler)
	}
	if !room.gamemetadata.gamestarted && !room.gamemetadata.gameended {
		if len(room.clients) < room.playerlimit {
			c.room = room
			c.ingame = true
			room.register <- c
		} else {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Target:  room.name,
				Message: fmt.Sprintf("oops the room %v is occupied :-P", room.name),
			}
			msg := message.encode()
			c.wsServer.broadcastToClient <- map[*client][]byte{c: msg}
		}
	} else {
		message := &Message{
			Action:  FailJoinRoomNotification,
			Target:  room.name,
			Message: fmt.Sprintf("the game is already started in %vs :-P", room.name),
		}
		msg := message.encode()
		c.wsServer.broadcastToClient <- map[*client][]byte{c: msg}
	}
}

func (c *client) HandleSendMessageToRoom(roomname string, message []byte) {
	// The send-message action, this will send messages to a specific room now.
	// Which room wil depend on the message Target
	if room := c.wsServer.findRoom(roomname); room != nil {
		room.broadcast <- message
	} else {
		log.Println("room not found")
	}
}

func (c *client) HandleSendMessageToBotOfRoom(roomname string, message []byte) {
	// The send-message action, this will send messages to a specific room now.
	// Which room wil depend on the message Target
	if room := c.wsServer.findRoom(roomname); room != nil {
		room.broadcastToBot <- message
	} else {
		log.Println("room not found")
	}
}
func (c *client) HandleSendMessageToClientInRoom(roomname string, clientids []string, message []byte) {
	c.wsServer.RLock()
	defer c.wsServer.RUnlock()
	if room := c.wsServer.findRoom(roomname); room != nil {
		clientmessage := make(map[string][]byte)
		for _, cid := range clientids {
			if val, ok := room.clients[cid]; val && ok {
				clientmessage[cid] = message
			} else {
				log.Println("client not found in a room")
			}
		}
		if len(clientmessage) > 0 {
			room.broadcastToClientinRoom <- clientmessage
		}
	} else {
		log.Println("room not found")
	}

}

func (c *client) HandleSendMessageToClients(clientids []string, message []byte) {
	lobbyServer := c.wsServer
	lobbyServer.RLock()
	defer lobbyServer.RUnlock()
	clientmessage := make(map[*client][]byte)
	for _, cid := range clientids {
		if client, ok := lobbyServer.clients[cid]; ok {
			clientmessage[client] = message
		} else {
			log.Println("client not found in a room")
		}
	}
	if len(clientmessage) > 0 {
		lobbyServer.broadcastToClient <- clientmessage
	}

}

func (c *client) HandleJoinRandomRoomMessage(roomname string) {
	c.wsServer.Lock()
	defer c.wsServer.Unlock()
	var room *room
	server := c.wsServer
	for _, findroom := range server.rooms {
		if len(findroom.clients) < findroom.playerlimit && findroom.randomgame && !findroom.gamemetadata.gamestarted && !findroom.gamemetadata.gameended {
			room = findroom
			log.Println("found random room", len(findroom.clients))
			break

		}
	}
	if room == nil {
		log.Println("Not found a random room creating a new one:-", roomname)
		if len(roomname) != 10 {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Message: fmt.Sprintf("Room name %v is not valid ", roomname),
			}
			msg := message.encode()
			c.wsServer.broadcastToClient <- map[*client][]byte{c: msg}
		}

		room = c.wsServer.createRoom(roomname, c, 2, true, c.wsServer.roomStateHandler)

	}
	if !room.gamemetadata.gamestarted && !room.gamemetadata.gameended {
		log.Println("Found random room :-", room.name)
		c.room = room
		c.ingame = true
		room.register <- c
	}
}

// HandleLeaveRoomMessage will handle leaving a room.unregister <- Client
func (client *client) HandleLeaveRoomMessage() {

	client.room.unregister <- client
}

// handleStartGameMessage will handle starting the game  a room.broadcast <- Message
// func (c *client) handleStartGameMessage(message Message) {
// 	log.Println("start the game")
// 	if room := c.wsServer.findRoom(message.Message); room != nil {
// 		botnotify := &Message{ //imp will change ui
// 			Action: StartGameNotification,
// 			Target: room.name,
// 		}
// 		room.gamemetadata.gamestarted = true //game started
// 		room.broadcast <- botnotify.encode() //bot greeting message posted

// 	}

// }
