package websocket

// processes requests
// tracks client connections
// handles subscriptions

import (
	"encoding/json"
	"go-chat-app-api/internal/auth"
	cmap "go-chat-app-api/internal/concurrent-map"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/presence"
	"unsafe"

	"github.com/nats-io/nats.go"
)

const (
	HUB_MAX_CLIENTS_PER_USER = 12
)

type WsHandler func(*WsHub, *WsServerEvent)

// websockets state(general)
type WsHub struct {
	mongoDBInst *database.MongoDBInstance
	auth        auth.Auth
	NATSConn    *nats.Conn

	register   chan *WsClient
	unregister chan *WsClient
	events     chan WsServerEvent

	clients           cmap.ConcurrentMap[*WsClient, bool]
	clientsByUid      cmap.ConcurrentMap[string, []*WsClient]
	statusSubscribers cmap.ConcurrentMap[string, map[*WsClient]bool]

	handlers map[int]WsHandler // read only, so can use plain map
}

// https://stackoverflow.com/a/57556517
func xorshift(n uint64, i uint) uint64 {
	return n ^ (n >> i)
}
func hash(n uint64) uint64 {
	var p uint64 = 0x5555555555555555   // pattern of alternating 0 and 1
	var c uint64 = 17316035218449499591 // random uneven integer constant;
	return c * xorshift(p*xorshift(n, 32), 32)
}

func NewWsHub(auth auth.Auth, mongoDBInst *database.MongoDBInstance, natsConn *nats.Conn) WsHub {
	// use custom sharding function because default one will do a lot of redudant ops(convert to string/alloc->hash it char by char)
	// TODO: remove this map?
	clients := cmap.NewWithCustomShardingFunction[*WsClient, bool](func(key *WsClient) uint32 {
		return uint32((hash(uint64(uintptr(unsafe.Pointer(key))))) % 32)
	})
	clientsByUid := cmap.New[[]*WsClient]()
	statusSubscribers := cmap.New[map[*WsClient]bool]()
	hub := WsHub{
		auth:              auth,
		mongoDBInst:       mongoDBInst,
		NATSConn:          natsConn,
		register:          make(chan *WsClient, 32),
		unregister:        make(chan *WsClient, 32),
		events:            make(chan WsServerEvent, 64),
		clients:           clients,
		clientsByUid:      clientsByUid,
		statusSubscribers: statusSubscribers,
		handlers:          make(map[int]WsHandler),
	}
	hub.handlers[WsEvent_SendMessageRequest] = handleSendMessage
	hub.handlers[WsEvent_OnlineStatusSubscribeRequest] = handleSubscribeOnlineStatus

	natsConn.Subscribe(NATSWsEventBroadcast, func(msg *nats.Msg) {
		event := WsEvent{}
		eventBytes := msg.Data
		err := json.Unmarshal(eventBytes, &event)
		if err != nil {
		}

		clients, ok := hub.clientsByUid.Get(event.To)
		if !ok {
			return
		}
		for _, c := range clients {
			c.responses <- event
		}
	})

	return hub
}

func (hub *WsHub) getClients(uid string) []*WsClient {
	res, ok := hub.clientsByUid.Get(uid)
	if !ok {
		return nil
	}

	return res
}
func (hub *WsHub) isUserOnline(uid string) bool {
	return len(hub.getClients(uid)) != 0
}
func (hub *WsHub) notifyClients(uid string, e WsEvent) {
	clients := hub.getClients(uid)

	for _, cl := range clients {
		cl.responses <- e
	}
}
func (hub *WsHub) Run() {
	for {
		select {
		case c := <-hub.register:
			{
				hub.clients.Set(c, true)

				hub.clientsByUid.Upsert(c.uid, nil, func(exist bool, valueInMap, newValue []*WsClient) []*WsClient {
					v := make([]*WsClient, len(valueInMap))
					copy(v, valueInMap)
					return append(v, c)
				})

				hub.NATSConn.Publish(presence.NATSNewChatUserConn, []byte(c.uid))
			}
		case c := <-hub.unregister:
			{
				c.unsubsribeAll(hub)
				hub.clients.Remove(c)
				hub.clientsByUid.Upsert(c.uid, nil, func(exist bool, valueInMap, newValue []*WsClient) []*WsClient {
					res := make([]*WsClient, 0, len(valueInMap)) // clone for safe get
					for _, cl := range valueInMap {
						if cl != c {
							res = append(res, cl)
						}
					}
					return res
				})

				hub.NATSConn.Publish(presence.NATSLostChatUserConn, []byte(c.uid))
			}
		case event := <-hub.events:
			{
				h, ok := hub.handlers[event.event.EventType]
				if !ok {
					continue
				}
				go h(hub, &event)
			}
		}
	}
}
