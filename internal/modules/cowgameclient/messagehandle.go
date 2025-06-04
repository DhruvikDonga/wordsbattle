package cowgameclient

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"unicode"
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
		client.handleSendMessage(message)
	case SendMessageActionByBot:
		client.handleBotMessage(message)
	case JoinRoomAction:
		client.handleJoinRoomMessage(message)
	case JoinRandomRoomAction:
		client.handleJoinRandomRoomMessage(message)
	case LeaveRoomAction:
		client.handleLeaveRoomMessage()
	case ClientNameAction:
		client.handleClientNameAction(message)
	case StartGameAction:
		client.handleStartGameMessage(message)
	}
}

func (client *Client) handleSendMessage(message Message) {
	// The send-message action, this will send messages to a specific room now.
	// Which room wil depend on the message Target
	roomname := message.Target
	if room := client.wsServer.findRoom(roomname); room != nil {
		room.broadcast <- &message
	}
}

func (client *Client) handleBotMessage(message Message) {
	//client requested bot to send message this is new word or previous word
	//clients and score
	roomname := message.Target

	if room := client.wsServer.findRoom(roomname); room != nil {
		log.Println("Message to the bot:-", message.Message)
		if strings.TrimSpace(message.Message) != "" { //user has a word
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
}

func (client *Client) handleClientNameAction(message Message) {
	clientname := strings.TrimSpace(message.Message)
	if clientname != "" {
		client.Name = clientname
	}
}

func isValidRoomName(roomname string) bool {
	if len(roomname) != 10 {
		return false
	}

	for _, r := range roomname {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// handleJoinRoomMessage will handle joining a room.register <- Client
func (client *Client) handleJoinRoomMessage(message Message) {
	roomname := strings.TrimSpace(message.Message)

	room := client.wsServer.findRoom(roomname) //this will be usefull is player is playing with a friend
	// room := client.wsServe.findidealroom() //go through list of ideal rooms  first then create a new random room
	if room == nil {
		if !isValidRoomName(roomname) {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Target:  roomname,
				Message: fmt.Sprintf("Room name %v is not valid ", roomname),
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
			return
		}
		room = client.wsServer.createRoom(roomname, client, 10, false)
		if room == nil {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Message: "Failed to create room",
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
			return
		}
	}

	if room.gamemetadata.gamestarted || room.gamemetadata.gameended {
		client.wsServer.broadcastToClient <- map[*Client]*Message{client: {
			Action:  FailJoinRoomNotification,
			Target:  room.name,
			Message: fmt.Sprintf("the game is already started in %vs :-P", room.name),
		}}
		return
	}

	if len(room.clients) >= room.playerlimit {
		client.wsServer.broadcastToClient <- map[*Client]*Message{client: {
			Action:  FailJoinRoomNotification,
			Target:  room.name,
			Message: fmt.Sprintf("oops the room %v is occupied :-P", room.name),
		}}
		return
	}

	client.room = room
	client.ingame = true
	room.register <- client
}

func (client *Client) handleJoinRandomRoomMessage(message Message) {
	roomname := strings.TrimSpace(message.Message)
	room := client.findAvailableRandomRoom()
	if room == nil {
		if !isValidRoomName(roomname) {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Message: fmt.Sprintf("Room name %v is not valid ", roomname),
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
		}

		log.Println("Not found a random room creating a new one:-", roomname)
		room = client.wsServer.createRoom(roomname, client, 2, true)
		if room == nil {
			message := &Message{
				Action:  FailJoinRoomNotification,
				Message: "Failed to create room",
			}
			client.wsServer.broadcastToClient <- map[*Client]*Message{client: message}
			return
		}
	}

	if room.gamemetadata.gamestarted || room.gamemetadata.gameended {
		log.Printf("Room %s game already in progress", room.name)
		return
	}

	log.Println("Joining random room:-", room.name)
	client.room = room
	client.ingame = true
	room.register <- client
}

func (client *Client) findAvailableRandomRoom() *Room {
	server := client.wsServer

	for room := range server.rooms {
		if len(room.clients) < room.playerlimit && room.randomgame && !room.gamemetadata.gamestarted && !room.gamemetadata.gameended {
			log.Println("found random room", len(room.clients))
			return room
		}
	}
	return nil
}

// handleLeaveRoomMessage will handle leaving a room.unregister <- Client
func (client *Client) handleLeaveRoomMessage() {
	if client.room == nil {
		return
	}

	select {
	case client.room.unregister <- client:
		log.Printf("Client %s left room %s", client.Name, client.room.name)
	default:
		log.Printf("Failed to unregister client %s from room %s", client.Name, client.room.name)
	}

	client.room = nil
	client.ingame = false
}

// handleStartGameMessage will handle starting the game  a room.broadcast <- Message
func (client *Client) handleStartGameMessage(message Message) {
	roomname := strings.TrimSpace(message.Message)
	room := client.wsServer.findRoom(roomname)
	if room == nil {
		log.Printf("Room %s not found for game start", roomname)
		return
	}

	if room.gamemetadata.gamestarted || room.gamemetadata.gameended {
		log.Printf("Game in room %s already started/ended", roomname)
		return
	}

	if len(room.clients) < 1 {
		log.Printf("Not enough players in room %s to start game", roomname)
		return
	}

	log.Println("Starting the game in room:", roomname)

	botnotify := &Message{
		Action: StartGameNotification,
		Target: room.name,
		Sender: &Client{Name: "bot-of-the-room"},
	}

	room.gamemetadata.gamestarted = true
	room.broadcast <- botnotify
}
