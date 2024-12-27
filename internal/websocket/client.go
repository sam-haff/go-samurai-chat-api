package websocket

import (
	"encoding/json"
	"fmt"
	cmap "go-chat-app-api/internal/concurrent-map"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// websocket state(per connection)
type WsClient struct {
	uid              string
	authToken        string
	responses        chan WsEvent
	subscribedTo     map[string]bool
	subscribedToLock sync.Mutex // cmap.ConcurrentMap is very bad for iteration

	conn *websocket.Conn
	hub  *WsHub
}

// clears structures for tracking subscribtions on special events for the client
func (c *WsClient) unsubsribeAll(hub *WsHub) {
	uidsPerShard := make([][]string, cmap.SHARD_COUNT)

	c.subscribedToLock.Lock()
	uids := make([]string, len(c.subscribedTo))
	i := 0
	for k := range c.subscribedTo {
		//exp
		shardIdx := hub.statusSubscribers.GetShardIndex(k)
		if uidsPerShard[shardIdx] == nil {
			uidsPerShard[shardIdx] = make([]string, 0, 16)
		}
		uidsPerShard[shardIdx] = append(uidsPerShard[shardIdx], k)
		//exp
		uids[i] = k
		i++
	}
	c.subscribedTo = nil
	c.subscribedToLock.Unlock()

	for shardIdx, uids := range uidsPerShard {
		shard := hub.statusSubscribers.GetShardByIndex(uint(shardIdx))
		shard.Lock()
		for _, uid := range uids {
			val, ok := shard.UnsafeGet(uid)
			if !ok {
				continue
			}

			res := make(map[*WsClient]bool, len(val))
			for k, v := range val {
				res[k] = v
			}

			shard.UnsafeSet(uid, res)
		}
		shard.Unlock()
	}

	/*
		// lots of same locks here
		for _, uid := range uids {
			hub.statusSubscribers.Upsert(uid, nil, func(exist bool, valueInMap, newValue map[*WsClient]bool) map[*WsClient]bool {
				if !exist || valueInMap == nil {
					return nil
				}
				res := make(map[*WsClient]bool, len(valueInMap)) // clone for safe get
				for k, v := range valueInMap {
					if k != c {
						res[k] = v
					}
				}
				return valueInMap
			})
		}*/
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

// read events/requests(WsEvent) from the client and send them(WsServerEvent) the hub events channel for further processing
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

// send request responses and events from the <responses> channel to the client
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
// upgrades the connection from http to websocket and
// launches threads for handling bidirectional communication.
func clientServeWs(hub *WsHub, w http.ResponseWriter, r *http.Request, token string, uid string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print("New client failed. " + err.Error() + "\n")
		return
	}

	client := &WsClient{hub: hub, conn: conn, authToken: token, uid: uid, responses: make(chan WsEvent, 32), subscribedTo: make(map[string]bool, 12)} // TODO: save auth token???
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
