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
		comm.AbortBadRequest(ctx, "Internal error", comm.CodeInvalidArgs)
		return
	}

	hub := val.(*WsHub)
	authToken, err := hub.auth.VerifyToken(ctx, params.Token) // TODO: move to middleware?
	if err != nil {
		comm.AbortBadRequest(ctx, "Not authenticated", comm.CodeNotAuthenticated)
		return
	}

	if len(hub.getClients(authToken.UID)) > HUB_MAX_CLIENTS_PER_USER {
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
		WsEvent_SendMessageRequest,
		true,
		e.event.Id,
		*comm.NewApiResponseWithJson("Success", comm.CodeSuccess, msg),
	)

	targetClients := hub.getClients(msg.ToId)
	senderClients := hub.getClients(e.origin.uid)
	clients := make([]*WsClient, 0, len(targetClients)+len(senderClients))
	if targetClients != nil {
		clients = append(clients, targetClients...)
	}
	if (msg.ToId != e.origin.uid) && (senderClients != nil) {
		clients = append(clients, senderClients...)
	}

	for _, cl := range clients {
		cl.responses <- commNewWsEventJSON(
			WsEvent_NewMessageEvent,
			false,
			e.event.Id,
			*comm.NewApiResponseWithJson("New message", comm.CodeSuccess, msg),
		)
	}
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

	isOnline := hub.isUserOnline(params.Uid)
	hub.notifyClients(e.origin.uid, NewOnlineStatusChangeEvent(params.Uid, isOnline))

	hub.statusSubscribers.Upsert(params.Uid, nil, func(exist bool, valueInMap, newValue map[*WsClient]bool) map[*WsClient]bool {
		if !exist || valueInMap == nil {
			m := make(map[*WsClient]bool)
			m[e.origin] = true
			return m
		}
		m := make(map[*WsClient]bool, len(valueInMap)) // clone for safe get
		for k, v := range valueInMap {
			m[k] = v
		}
		m[e.origin] = true
		return m
	})
	e.origin.subscribedToLock.Lock()
	if e.origin.subscribedTo == nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_OnlineStatusSubscribeRequest,
			true,
			e.event.Id,
			*comm.NewApiResponse("Client state is invalid", comm.CodeInvalidArgs),
		)
		e.origin.subscribedToLock.Unlock()
		return
	}
	e.origin.subscribedTo[params.Uid] = true
	e.origin.subscribedToLock.Unlock()

	e.origin.responses <- commNewWsEvent(
		WsEvent_OnlineStatusSubscribeRequest,
		true,
		e.event.Id,
		*comm.NewApiResponse("Subscribed", comm.CodeSuccess),
	)
}

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
		WsEvent_SendMessageRequest,
		true,
		e.event.Id,
		*comm.NewApiResponseWithJson("Success", comm.CodeSuccess, isOnlineRespObj),
	)
}
