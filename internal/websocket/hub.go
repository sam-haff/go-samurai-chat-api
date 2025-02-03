package websocket

// processes requests
// tracks client connections
// handles subscriptions

import (
	"go-chat-app-api/internal/auth"
	cmap "go-chat-app-api/internal/concurrent-map"
	"go-chat-app-api/internal/database"
	"unsafe"
)

const (
	HUB_MAX_CLIENTS_PER_USER = 12
)

type WsHandler func(*WsHub, *WsServerEvent)

// websockets state(general)
type WsHub struct {
	mongoDBInst *database.MongoDBInstance
	auth        auth.Auth

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

func NewWsHub(auth auth.Auth, mongoDBInst *database.MongoDBInstance) WsHub {
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

				res := hub.clientsByUid.Upsert(c.uid, nil, func(exist bool, valueInMap, newValue []*WsClient) []*WsClient {
					v := make([]*WsClient, len(valueInMap))
					copy(v, valueInMap)
					return append(v, c)
				})
				if len(res) == 1 {
					// first user's client is connected -> user is online now

					subs, _ := hub.statusSubscribers.Get(c.uid)
					for sub, _ := range subs {
						hub.notifyClients(sub.uid, NewOnlineStatusChangeEvent(c.uid, true))
					}
				}
			}
		case c := <-hub.unregister:
			{
				c.unsubsribeAll(hub)
				hub.clients.Remove(c)
				res := hub.clientsByUid.Upsert(c.uid, nil, func(exist bool, valueInMap, newValue []*WsClient) []*WsClient {
					res := make([]*WsClient, 0, len(valueInMap)) // clone for safe get
					for _, cl := range valueInMap {
						if cl != c {
							res = append(res, cl)
						}
					}
					return res
				})
				// last user's client disconnected -> offline
				if len(res) == 0 {
					subs, _ := hub.statusSubscribers.Get(c.uid)
					for sub := range subs {
						hub.notifyClients(sub.uid, NewOnlineStatusChangeEvent(c.uid, false))
					}
				}
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
