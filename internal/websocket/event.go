package websocket

// primitive of websocket communication

import "encoding/json"

const (
	WsEvent_NewMessageEvent              = 0
	WsEvent_SendMessageRequest           = 1
	WsEvent_CheckOnlineRequest           = 2
	WsEvent_OnlineStatusChangeEvent      = 3
	WsEvent_OnlineStatusSubscribeRequest = 4
)

type WsEvent struct {
	EventType  int             `json:"event_type"`
	Id         string          `json:"id"` //uuid
	Obj        json.RawMessage `json:"obj"`
	IsResponse bool            `json:"is_response"`
}

type WsServerEvent struct {
	origin *WsClient
	event  WsEvent
}

type WsOnlineStatus struct {
	Uid    string `json:"uid"`
	Online bool   `json:"online"`
}

func NewOnlineStatusChangeEvent(targetUid string, online bool) WsEvent {
	obj := WsOnlineStatus{
		Uid:    targetUid,
		Online: online,
	}
	b, _ := json.Marshal(&obj)

	return WsEvent{
		EventType:  WsEvent_OnlineStatusChangeEvent,
		Id:         "", // TODO: ???
		IsResponse: false,
		Obj:        b,
	}
}
