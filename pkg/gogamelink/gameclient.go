package gogamelink

//Frontend[ws->write] ----> websocket pipe ---> Go(WSClient readpump <-broadcast channel) ---> (Websocket server(channel broadcast) ---> broadcast channel and broadcasttoAllClient(for loop <-send(channel of writepump part of client)) --> Go(WSClient writepump (to send over other client) --> /ws message --> frontend of other client ---> readmessage buffer and back to first

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
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

type ClientStorer interface {
	/**HandleMessages evokes in readpump() of client**/
	/**Create this function to do the handling of the custom message struct**/
	/**Handle the custom json messages**/
	HandleMessage(c Client, jsonmessage []byte, clientID string, roomID string)
}

type Client interface {
	//AddClientToRoom handles to joining of client to room
	//add it in HandleMessage to handle the room join
	//takes roomname in its argument
	AddClientToRoom(roomname string)

	//HandleJoinRandomRoomMessage handles to joining of client to a random room
	//add it in HandleMessage to handle the join random room
	//takes roomname in its argument from frontend incase if no randomroom found than create a room
	HandleJoinRandomRoomMessage(roomname string)

	//HandleLeaveRoomMessage handles to leaving of client from a room
	//add it in HandleMessage to handle the room leave
	HandleLeaveRoomMessage()

	//HandleSendMessageToRoom send message to all the clients present in a room
	//takes roomname of type string and message of []byte (json)
	HandleSendMessageToRoom(roomname string, message []byte)

	//HandleSendMessageToBotOfRoom send message to roomhandler function which you have created
	//takes roomname of type string and message of []byte (json)
	HandleSendMessageToBotOfRoom(roomname string, message []byte)

	//HandleSendMessageToClientInRoom sends message to a particular clients in a room
	//takes roomname , clientids of type []string and message of  []byte (json)
	HandleSendMessageToClientInRoom(roomname string, clientids []string, message []byte)

	//HandleSendMessageToClients sends message to a particular clients which are in lobbyserver it does not use roomid
	//takes clientids of type []string and message of  []byte (json)
	/**Use it for cases example failure messages , notifications to a client if incase its not in room**/
	HandleSendMessageToClients(clientids []string, message []byte)
}

// Client represents the websocket client at the server [server] <--> [client]
type client struct {
	Slug string `json:"slug"` //unique big string to avoid confusion
	// websocket connection
	conn     *websocket.Conn
	wsServer *LobbyServer //keep reference of webserver to every client
	send     chan []byte  //send channel then moves this to broadcast channel in server
	room     *room        //a client will be in one room at a time
	ingame   bool         //is in any game
	Storer   ClientStorer
}

// newClient initialize new websocket client like App server in routes.go
func newClient(wscon *websocket.Conn, wsServer *LobbyServer, name string, customhandler ClientStorer) *client {
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
	c := &client{
		Slug:     randomname,
		conn:     wscon,
		wsServer: wsServer,
		send:     make(chan []byte, 256), //needs to be buffered cause it should not block when channel is not receiving from broadcast
		room:     &room{},
		ingame:   false, //not in any game
		Storer:   customhandler,
	}
	wsServer.register <- c //we are registering the client
	return c
}

func (c *client) GetClientSlug() string {
	return c.Slug
}

// readPump Goroutine, the client will read new messages send over the WebSocket connection. It will do so in an endless loop until the client is disconnected. When the connection is closed, the client will call its own disconnect method to clean up.
// upon receiving new messages the client will push them in the LobbyServer broadcast channel.
func (client *client) readPump() {
	defer func() {
		client.Disconnect()
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
			if _, ok := room.clients[client.Slug]; ok {
				delete(room.clients, client.Slug)
				if len(room.clients) == 0 {
					room.stoproom <- true
				}
			}

			break
		}

		log.Println("JSON-MESSAGE-readpump", string(jsonMessage))
		//client.wsServer.broadcast <- jsonMessage //not to brodcast all brodcast to rooms
		clientID := client.Slug
		var roomID string
		if client.room != nil {
			roomID = client.room.name
		} else {
			roomID = "global"
		}
		client.Storer.HandleMessage(client, jsonMessage, clientID, roomID)
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

		room = c.wsServer.CreateRoom(roomname, c, 2, true, c.wsServer.roomStateHandler)

	}
	if !room.gamemetadata.gamestarted && !room.gamemetadata.gameended {
		log.Println("Found random room :-", room.name)
		c.room = room
		c.ingame = true
		room.register <- c
	}
}

// writePump goroutine handles sending the messages to the connected client. It runs in an endless loop waiting for new messages in the client.send channel. When receiving new messages it writes them to the client, if there are multiple messages available they will be combined in one write.
// writePump is also responsible for keeping the connection alive by sending ping messages to the client with the interval given in pingPeriod. If the client does not respond with a pong, the connection is closed.
func (client *client) writePump() {
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

func (client *client) Disconnect() {
	log.Println("disconnect", client.Slug)
	client.wsServer.unregister <- client //remove client from webserver map list
	if _, ok := client.wsServer.rooms[client.room.name]; ok {
		client.room.unregister <- client //unregister to game room
	}
	close(client.send)  //close the sending channel
	client.conn.Close() //close the client connection
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *LobbyServer, w http.ResponseWriter, r *http.Request, messagehandler ClientStorer) {

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Println(err)
		return
	}

	name := r.URL.Query().Get("name") // ws://url?name=dumm_name

	if len(name) < 1 {
		name = "Guest"
	}

	client := newClient(conn, wsServer, name, messagehandler)

	log.Println("New client joined the hub")
	log.Println(client)

	//log.Println(client.Gamemetadata.Color)
	go client.readPump()
	go client.writePump()

}
