package cowgameclient

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// handleNewMessage will handle Client messages
func (client *Client) handleNewMessage(jsonMessage []byte) {
	var message Message
	if err := json.Unmarshal(jsonMessage, &message); err != nil {
		log.Printf("Error on unmarshal JSON message %s", err)
	}

	message.Sender = client //attach client object as sender of message

	switch message.Action {
	case SendMessageAction:
		// The send-message action, this will send messages to a specific room now.
		// Which room wil depend on the message Target
		roomname := message.Target
		if room := client.wsServer.findRoom(roomname); room != nil {
			room.broadcast <- &message
		}

	case SendMessageActionByBot: //client requested bot to send message this is new word or previous word
		//clients and score
		roomname := message.Target

		if room := client.wsServer.findRoom(roomname); room != nil {
			log.Println("Message to the bot:-", message.Message)
			if message.Message != "" { //user has a word
				wordofusermessage := &Message{
					Target:  room.name,
					Message: message.Message,
					Action:  SendMessageAction,
					Sender:  message.Sender,
				}
				room.gamemetadata.wordguessed = message.Message
				log.Println(room.gamemetadata.gameturn)
				room.gamemetadata.gameturn = true //ticker will change game state
				room.broadcast <- wordofusermessage
			}
		}

	case JoinRoomAction:
		client.handleJoinRoomMessage(message)

	case JoinRandomRoomAction:
		client.handleJoinRandomRoomMessage(message)

	case LeaveRoomAction:
		client.handleLeaveRoomMessage(message)
	case ClientNameAction:
		clientname := message.Message
		client.Name = clientname

	case StartGameAction:
		client.handleStartGameMessage(message)

	}
}

// handleJoinRoomMessage will handle joining a room.register <- Client
func (client *Client) handleJoinRoomMessage(message Message) {
	roomname := message.Message
	room := client.wsServer.findRoom(roomname) //this will be usefull is player is playing with a friend
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
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
			return
		}
		room = client.wsServer.createRoom(roomname, client, 10, false)
	}
	if room.gamemetadata.gamestarted == false && room.gamemetadata.gameended == false {
		if len(room.clients) < room.playerlimit {
			client.room = room
			client.ingame = true
			room.register <- client
		} else {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Target:  room.name,
				Message: fmt.Sprintf("oops the room %v is occupied :-P", room.name),
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
		}
	} else {
		message := &Message{
			Action:  FailJoinRoomNotification,
			Target:  room.name,
			Message: fmt.Sprintf("the game is already started in %vs :-P", room.name),
		}
		client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
	}
}

func (client *Client) handleJoinRandomRoomMessage(message Message) {
	var room *Room
	server := client.wsServer
	for findroom := range server.rooms {
		if len(findroom.clients) < findroom.playerlimit && findroom.randomgame == true && findroom.gamemetadata.gamestarted == false && findroom.gamemetadata.gameended == false {
			room = findroom
			log.Println("found random room", len(findroom.clients))
			break

		}
	}
	if room == nil {
		log.Println("Not found a random room creating a new one:-", message.Message)
		roomname := message.Message
		if len(roomname) != 10 {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Message: fmt.Sprintf("Room name %v is not valid ", roomname),
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
		}

		room = client.wsServer.createRoom(roomname, client, 2, true)

	}
	if room.gamemetadata.gamestarted == false && room.gamemetadata.gameended == false {

		log.Println("Found random room :-", room.name)
		client.room = room
		client.ingame = true
		room.register <- client

	}

}

// handleLeaveRoomMessage will handle leaving a room.unregister <- Client
func (client *Client) handleLeaveRoomMessage(message Message) {

	client.room.unregister <- client
}

// handleStartGameMessage will handle starting the game  a room.broadcast <- Message
func (client *Client) handleStartGameMessage(message Message) {
	log.Println("start the game")
	if room := client.wsServer.findRoom(message.Message); room != nil {
		botnotify := &Message{ //imp will change ui
			Action: StartGameNotification,
			Target: room.name,
			Sender: &Client{Name: "bot-of-the-room"},
		}
		room.gamemetadata.gamestarted = true //game started
		room.broadcast <- botnotify          //bot greeting message posted

	}

}
