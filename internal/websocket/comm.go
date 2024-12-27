package websocket

// communication helpers
// shortcuts for creation of various events

import (
	"encoding/json"
	"go-chat-app-api/internal/comm"
)

func commNewWsEvent(eventType int, isResponse bool, eventId string, resp comm.ApiResponsePlain) WsEvent {
	b, _ := json.Marshal(resp)

	return WsEvent{
		EventType:  eventType,
		IsResponse: isResponse,
		Id:         eventId,
		Obj:        b,
	}
}

func commNewWsEventJSON(eventType int, isResponse bool, eventId string, resp comm.ApiResponseWithJson) WsEvent {
	b, _ := json.Marshal(resp)

	return WsEvent{
		EventType:  eventType,
		IsResponse: isResponse,
		Id:         eventId,
		Obj:        b,
	}
}
