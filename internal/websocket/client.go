package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsClient struct {
	uid       string
	authToken string
	responses chan WsEvent

	conn *websocket.Conn
	hub  *WsHub
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (c *WsClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		event := WsEvent{}
		err = json.Unmarshal(message, &event)
		if err != nil {
			fmt.Printf("Failed to parse request: %s\n", err.Error())
			// TODO:???
			continue
		}
		c.hub.events <- WsServerEvent{
			origin: c,
			event:  event,
		}
	}
}

func (c *WsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case event, ok := <-c.responses:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			eventBytes, err := json.Marshal(event)
			if err != nil {
				continue
			}

			w.Write(eventBytes)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func clientServeWs(hub *WsHub, w http.ResponseWriter, r *http.Request, token string, uid string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print("New client failed. " + err.Error() + "\n")
		return
	}

	client := &WsClient{hub: hub, conn: conn, authToken: token, uid: uid, responses: make(chan WsEvent, 256)} // TODO: save auth token???
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
