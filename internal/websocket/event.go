package websocket

import "encoding/json"

const (
	WsEvent_NewMessageEvent     = 0
	WsEvent_SendMessageRequest  = 1
	WsEvent_SendMessageResponse = 2 // TODO: store bool to signal response(isResponse)
)

type WsEvent struct {
	EventType int             `json:"event_type"`
	Id        string          `json:"id"` //uuid
	Obj       json.RawMessage `json:"obj"`
}

type WsServerEvent struct {
	origin *WsClient
	event  WsEvent
}
