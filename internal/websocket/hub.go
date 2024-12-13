package websocket

import (
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
)

type WsHandler func(*WsHub, *WsServerEvent)

type WsHub struct {
	mongoDBInst *database.MongoDBInstance
	auth        auth.Auth

	register     chan *WsClient
	unregister   chan *WsClient
	events       chan WsServerEvent
	clients      map[*WsClient]bool
	clientsByUid map[string][]*WsClient

	handlers map[int]WsHandler
}

func NewWsHub(auth auth.Auth, mongoDBInst *database.MongoDBInstance) WsHub {
	hub := WsHub{
		auth:         auth,
		mongoDBInst:  mongoDBInst,
		register:     make(chan *WsClient, 32),
		unregister:   make(chan *WsClient, 32),
		events:       make(chan WsServerEvent, 64),
		clients:      make(map[*WsClient]bool, 2048),
		clientsByUid: make(map[string][]*WsClient, 2048),
		handlers:     make(map[int]WsHandler),
	}
	hub.handlers[WsEvent_SendMessageRequest] = handleSendMessage

	return hub
}

func (hub *WsHub) getClients(uid string) []*WsClient {
	res, ok := hub.clientsByUid[uid]
	if !ok {
		return nil
	}

	return res
}

func (hub *WsHub) Run() {
	for {
		select {
		case c := <-hub.register:
			{
				hub.clients[c] = true
				res, ok := hub.clientsByUid[c.uid]
				if !ok || res == nil {
					hub.clientsByUid[c.uid] = make([]*WsClient, 0)
				}
				hub.clientsByUid[c.uid] = append(hub.clientsByUid[c.uid], c)
			}
		case c := <-hub.unregister:
			{
				delete(hub.clients, c)
				res, ok := hub.clientsByUid[c.uid]
				if !ok || res == nil {
					continue
				}
				for i, cl := range res {
					if cl == c {
						res = append(res[:i], res[i+1:]...)
						hub.clientsByUid[c.uid] = res
						break
					}
				}
			}
		case event := <-hub.events:
			{
				h, ok := hub.handlers[event.event.EventType]
				if !ok {
					continue
				}
				h(hub, &event)
			}
		}
	}
}
