package gomeshstream

import (
	"encoding/json"
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

type client struct {
	slug string

	conn       *websocket.Conn
	meshServer *meshServer //keep reference of webserver to every client
	send       chan []byte
}

// newClient initialize new websocket client like App server in routes.go
func newClient(wscon *websocket.Conn, m *meshServer, name string) *client {
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
		slug:       randomname,
		conn:       wscon,
		meshServer: m,
		send:       make(chan []byte, 256), //needs to be buffered cause it should not block when channel is not receiving from broadcast
	}
	m.clientConnect <- c //we are registering the client
	return c
}

// readPump Goroutine, the client will read new messages send over the WebSocket connection. It will do so in an endless loop until the client is disconnected. When the connection is closed, the client will call its own disconnect method to clean up.
// upon receiving new messages the client will push them in the LobbyServer broadcast channel.
func (client *client) readPump() {
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
			break
		}

		var message Message
		if err := json.Unmarshal(jsonMessage, &message); err != nil {
			log.Printf("Error on unmarshal JSON message %s", err)
		}
		message.Sender = client.slug
		log.Println("JSON-MESSAGE-readpump", string(jsonMessage))
		client.meshServer.processMessage <- &message

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

func (client *client) disconnect() {
	m := client.meshServer
	m.clientDisconnect <- client
	close(client.send)  //close the sending channel
	client.conn.Close() //close the client connection
}

// ServeWs handles websocket requests from clients requests.
func ServeWs(meshserv *meshServer, w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Upgrade(w, r, nil, 4096, 4096)
	if err != nil {
		log.Println(err)
		return
	}

	name := r.URL.Query().Get("name") // ws://url?name=dumm_name

	if len(name) < 1 {
		name = "Guest"
	}

	client := newClient(conn, meshserv, name)

	log.Println("New client ", client.slug, " joined the hub")

	//log.Println(client.Gamemetadata.Color)
	go client.readPump()
	go client.writePump()

}
