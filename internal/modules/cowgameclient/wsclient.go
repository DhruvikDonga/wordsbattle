package cowgameclient

//Frontend[ws->write] ----> websocket pipe ---> Go(WSClient readpump <-broadcast channel) ---> (Websocket server(channel broadcast) ---> broadcast channel and broadcasttoAllClient(for loop <-send(channel of writepump part of client)) --> Go(WSClient writepump (to send over other client) --> /ws message --> frontend of other client ---> readmessage buffer and back to first

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
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

var (
	newline = []byte{'\n'}
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
	wsServer     *LobbyServer        //keep reference of webserver to every client
	send         chan []byte         //send channel then moves this to broadcast channel in server
	room         *Room               //a client will be in one room at a time
	ingame       bool                //is in any game
	Gamemetadata *ClientGameMetadata `json:"clientgamemetadata"`
	closeOnce    sync.Once
}

// NewClient initialize new websocket client like App server in routes.go
func NewClient(wscon *websocket.Conn, wsServer *LobbyServer, name string, slug string) *Client {
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

// readPump Goroutine, the client will read new messages send over the WebSocket connection. It will do so in an endless loop until the client is disconnected. When the connection is closed, the client will call its own disconnect method to clean up.
// upon receiving new messages the client will push them in the LobbyServer broadcast channel.
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
			if room != nil && room.clients != nil {
				room.mu.Lock()
				if _, ok := room.clients[client]; ok {
					delete(room.clients, client)
					if len(room.clients) == 0 {
						select {
						case room.stoproom <- true:
						default:
						}
					}
				}
				room.mu.Unlock()
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
		client.disconnect()
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
	if client.room != nil {
		select {
		case client.room.unregister <- client:
		default:
		}
	}
	client.closeOnce.Do(func() {
		close(client.send)
	}) //close the sending channel
	client.conn.Close() //close the client connection
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(wsServer *LobbyServer, w http.ResponseWriter, r *http.Request) {

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
