package websocket

import (
	"encoding/json"
	"go-chat-app-api/internal/comm"
)

func commNewWsEvent(eventType int, eventId string, resp comm.ApiResponsePlain) WsEvent {
	b, _ := json.Marshal(resp)

	return WsEvent{
		EventType: eventType,
		Id:        eventId,
		Obj:       b,
	}
}

func commNewWsEventJSON(eventType int, eventId string, resp comm.ApiResponseWithJson) WsEvent {
	b, _ := json.Marshal(resp)

	return WsEvent{
		EventType: eventType,
		Id:        eventId,
		Obj:       b,
	}
}
