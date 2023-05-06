package handler

//Frontend[ws->write] ----> websocket pipe ---> Go(WSClient readpump <-broadcast channel) ---> (Websocket server(channel broadcast) ---> broadcast channel and broadcasttoAllClient(for loop <-send(channel of writepump part of client)) --> Go(WSClient writepump (to send over other client) --> /ws message --> frontend of other client ---> readmessage buffer and back to first

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const charintset = "0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type ClientGameMetadata struct { //some properties constraint to a room
	Color string `json:"color"` //color or avatar in room
	Score int    `json:"score"` //score in a room
}

// Client represents the websocket client at the server [server] <--> [client]
type Client struct {
	Name string `json:"name"` //name got from UI
	Slug string `json:"slug"` //unique big string to avoid confusion
	// websocket connection
	conn         *websocket.Conn
	wsServer     *WsServer           //keep reference of webserver to every client
	send         chan []byte         //send channel then moves this to broadcast channel in server
	room         *Room               //a client will be in one room at a time
	ingame       bool                //is in any game
	Gamemetadata *ClientGameMetadata `json:"clientgamemetadata"`
}

// NewClient initialize new websocket client like App server in routes.go
func NewClient(wscon *websocket.Conn, wsServer *WsServer, name string, slug string) *Client {
	return &Client{
		Name:         name,
		Slug:         slug,
		conn:         wscon,
		wsServer:     wsServer,
		send:         make(chan []byte, 256), //needs to be buffered cause it should not block when channel is not receiving from broadcast
		room:         &Room{},
		ingame:       false,                 //not in any game
		Gamemetadata: &ClientGameMetadata{}, //will count if he is in game
	}
}

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

// readPump Goroutine, the client will read new messages send over the WebSocket connection. It will do so in an endless loop until the client is disconnected. When the connection is closed, the client will call its own disconnect method to clean up.
// upon receiving new messages the client will push them in the WsServer broadcast channel.
func (client *Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.conn.SetReadLimit(maxMessageSize)
	// Frontend client will give a pong message to the routine we have to handle it
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(appData string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for message from client
	for {
		_, jsonMessage, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("\nunexepected close error: %v", err)
				break
			}
			room := client.room
			if _, ok := room.clients[client]; ok {
				delete(room.clients, client)
				if len(room.clients) == 0 {
					room.stoproom <- true
				}
			}

			break
		}
		isMessage := map[string]string{}

		json.Unmarshal(jsonMessage, &isMessage)
		log.Println(isMessage)
		log.Println("JSON-MESSAGE-readpump", string(jsonMessage))
		//client.wsServer.broadcast <- jsonMessage //not to brodcast all brodcast to rooms

		client.handleNewMessage(jsonMessage) //broadcat to room

	}
}

// writePump goroutine handles sending the messages to the connected client. It runs in an endless loop waiting for new messages in the client.send channel. When receiving new messages it writes them to the client, if there are multiple messages available they will be combined in one write.
// writePump is also responsible for keeping the connection alive by sending ping messages to the client with the interval given in pingPeriod. If the client does not respond with a pong, the connection is closed.
func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // not ok means send channel has been closed caused by disconnect() in readPump()
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Println("Message send channel:-", string(message))
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Attach queued chat messages to the current websocket message.
			n := len(client.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-client.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C: //make a ping request
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (client *Client) disconnect() {
	log.Println("disconnect", client.Slug)
	client.wsServer.unregister <- client //remove client from webserver map list
	if _, ok := client.wsServer.rooms[client.room]; ok {
		client.room.unregister <- client //unregister to game room
	}
	close(client.send)  //close the sending channel
	client.conn.Close() //close the client connection
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Upgrade(w, r, nil, 4096, 4096)
	if err != nil {
		log.Println(err)
		return
	}
	//random name
	a := make([]byte, 5)
	for i := range a {
		a[i] = charset[seededRand.Intn(len(charset))]
	}
	b := make([]byte, 3)
	for i := range b {
		b[i] = charintset[seededRand.Intn(len(charintset))]
	}
	randomname := string(a) + string(b)

	name := r.URL.Query().Get("name") // ws://url?name=dumm_name

	if len(name) < 1 {
		name = "Guest"
	}

	client := NewClient(conn, wsServer, name, randomname)

	log.Println("New client joined the hub")
	log.Println(client)

	log.Println(client.Gamemetadata.Color)
	go client.readPump()
	go client.writePump()

	wsServer.register <- client //we are registering the client
}
