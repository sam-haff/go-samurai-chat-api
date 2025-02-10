package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/messages"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	CtxVarHub = "hub"
)

func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup, wsRoutes *gin.RouterGroup) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO: do actual checks?
		return true
	}
	wsRoutes.Any("/ws", handleWs)
}

type WsParams struct {
	Token string `form:"token"`
}

func handleWs(ctx *gin.Context) {
	params := WsParams{}
	if err := ctx.BindQuery(&params); err != nil {
		fmt.Printf("Failed to bind WS query\n")
		comm.AbortBadRequest(ctx, "Invalid args", comm.CodeInvalidArgs)
		return
	}
	val, ok := ctx.Get(CtxVarHub) // TODO: must get
	if !ok {
		fmt.Printf("Failed, no vars set \n")
		comm.AbortBadRequest(ctx, "Internal error", comm.CodeInvalidArgs)
		return
	}

	hub := val.(*WsHub)
	authToken, err := hub.auth.VerifyToken(ctx, params.Token) // TODO: move to middleware?
	if err != nil {
		comm.AbortUnauthorized(ctx, "Not authenticated", comm.CodeNotAuthenticated)
		return
	}

	if len(hub.getClients(authToken.UID)) > HUB_MAX_CLIENTS_PER_USER {
		fmt.Printf("Too much connections by client \n")
		comm.AbortBadRequest(ctx, "Reached max clients per user", comm.CodeInvalidArgs)
		return
	}

	//if accounts.DBUserRegisterCompletedUtil()

	clientServeWs(hub, ctx.Writer, ctx.Request, params.Token, authToken.UID) // TODO: rename func
}

func handleSendMessage(hub *WsHub, e *WsServerEvent) {
	params := messages.AddMessageParams{}
	objBytes, _ := e.event.Obj.MarshalJSON() // TODO: check err
	err := json.Unmarshal(objBytes, &params)

	if err != nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Ill formed request", comm.CodeInvalidArgs),
		)

		return
	}

	fromUid := e.origin.uid

	toUserData := accounts.UserData{}
	if status := accounts.DBGetUserDataUtil(context.TODO(), hub.mongoDBInst, params.ToId, &toUserData); status != accounts.UtilErrorOk {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Contact is not registered properly", comm.CodeInvalidArgs),
		)
		return
	}

	msg := messages.NewMessageData(fromUid, params.ToId, params.Msg)
	err = messages.DBAddMessageUtil(context.TODO(), hub.mongoDBInst, msg)
	if err != nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("DB error", comm.CodeInvalidArgs),
		)

		return
	}

	e.origin.responses <- commNewWsEventJSON(
		e.origin.uid,
		WsEvent_SendMessageRequest,
		true,
		e.event.Id,
		*comm.NewApiResponseWithJson("Success", comm.CodeSuccess, msg),
	)

	// TODO: send event for broadcasting using nats

	/*targetClients := hub.getClients(msg.ToId)
	senderClients := hub.getClients(e.origin.uid)
	clients := make([]*WsClient, 0, len(targetClients)+len(senderClients))
	if targetClients != nil {
		clients = append(clients, targetClients...)
	}
	if (msg.ToId != e.origin.uid) && (senderClients != nil) {
		clients = append(clients, senderClients...)
	}*/

	recipients := []string{msg.ToId, e.origin.uid}
	for _, uid := range recipients {
		event := commNewWsEventJSON(
			uid,
			WsEvent_NewMessageEvent,
			false,
			e.event.Id,
			*comm.NewApiResponseWithJson("New message", comm.CodeSuccess, msg),
		)
		eventBytes, _ := json.Marshal(event)
		hub.NATSConn.Publish(NATSWsEventBroadcast, eventBytes)

	}

	/*for _, cl := range clients {
		cl.responses <- commNewWsEventJSON(
			WsEvent_NewMessageEvent,
			false,
			e.event.Id,
			*comm.NewApiResponseWithJson("New message", comm.CodeSuccess, msg),
		)
	}*/
}

type CheckOnlineParams struct {
	Uid string `json:"uid" binding:"min=1,max=1024,required"`
}

func handleSubscribeOnlineStatus(hub *WsHub, e *WsServerEvent) {
	params := CheckOnlineParams{}
	b, _ := e.event.Obj.MarshalJSON()
	err := json.Unmarshal(b, &params)

	if err != nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_OnlineStatusSubscribeRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Ill formed request", comm.CodeInvalidArgs),
		)
		return
	}
	toUserData := accounts.UserData{}
	if status := accounts.DBGetUserDataUtil(context.TODO(), hub.mongoDBInst, params.Uid, &toUserData); status != accounts.UtilErrorOk {
		e.origin.responses <- commNewWsEvent(
			WsEvent_OnlineStatusSubscribeRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Contact is not registered properly", comm.CodeInvalidArgs),
		)
		return
	}

	e.origin.subscribeOnOnlineStatusNATS(params.Uid)

	isOnline := queryPresenceOnlineStatus(params.Uid)
	hub.notifyClients(e.origin.uid, NewOnlineStatusChangeEvent(e.origin.uid, params.Uid, isOnline)) // TODO: notify only caller
	//event := NewOnlineStatusChangeEvent(e.origin.uid, params.Uid, isOnline)
	//eventBytes, _ := json.Marshal(event)
	// Ideally, we need to only notify only the origin
	// Use case: 1 user sits on the 2 devices. Addes new contact from 1st one,
	// it is automatically added on the 2nd device(potentially, on another hub).
	// Ideally second device should then call subscribtion request and then he
	//hub.NATSConn.Publish("wsevent_broadcast", eventBytes)

	e.origin.responses <- commNewWsEvent(
		WsEvent_OnlineStatusSubscribeRequest,
		true,
		e.event.Id,
		*comm.NewApiResponse("Subscribed", comm.CodeSuccess),
	)
}

// Deprecated: uses only single instance of hub, doesn't query all the hubs.
// Now the way is to query <presence> service(/online/:uid route).
func handleCheckOnline(hub *WsHub, e *WsServerEvent) {
	params := CheckOnlineParams{}
	objBytes, _ := e.event.Obj.MarshalJSON() // TODO: check err
	err := json.Unmarshal(objBytes, &params)

	if err != nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_CheckOnlineRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Ill formed request", comm.CodeInvalidArgs),
		)

		return
	}

	toUserData := accounts.UserData{}
	if status := accounts.DBGetUserDataUtil(context.TODO(), hub.mongoDBInst, params.Uid, &toUserData); status != accounts.UtilErrorOk {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Contact is not registered properly", comm.CodeInvalidArgs),
		)

		return
	}

	isOnline := hub.isUserOnline(params.Uid)
	isOnlineRespObj := struct {
		IsOnline bool `json:"is_online"`
	}{isOnline}

	e.origin.responses <- commNewWsEventJSON(
		e.origin.uid,
		WsEvent_SendMessageRequest,
		true,
		e.event.Id,
		*comm.NewApiResponseWithJson("Success", comm.CodeSuccess, isOnlineRespObj),
	)
}
