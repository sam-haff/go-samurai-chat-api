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

	clientServeWs(hub, ctx.Writer, ctx.Request, params.Token, authToken.UID) // TODO: rename func
}

func handleSendMessage(hub *WsHub, e *WsServerEvent) {
	params := messages.AddMessageParams{}
	objBytes, _ := e.event.Obj.MarshalJSON() // TODO: check err
	err := json.Unmarshal(objBytes, &params)

	if err != nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageResponse,
			e.event.Id,
			*comm.NewApiResponse("Ill formed request", comm.CodeInvalidArgs),
		)

		return
	}

	fromUserData := accounts.UserData{}
	if status := accounts.DBGetUserDataUtil(context.TODO(), hub.mongoDBInst, e.origin.uid, &fromUserData); status != accounts.UtilErrorOk {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageResponse,
			e.event.Id,
			*comm.NewApiResponse("Sender is not registered properly", comm.CodeInvalidArgs),
		)

		return
	}
	toUserData := accounts.UserData{}
	if status := accounts.DBGetUserDataUtil(context.TODO(), hub.mongoDBInst, params.ToId, &toUserData); status != accounts.UtilErrorOk {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageResponse,
			e.event.Id,
			*comm.NewApiResponse("Contact is not registered properly", comm.CodeInvalidArgs),
		)

		return
	}

	msg := messages.NewMessageData(fromUserData, params.ToId, params.Msg)

	err = messages.DBAddMessageUtil(context.TODO(), hub.mongoDBInst, msg)
	if err != nil {
		e.origin.responses <- commNewWsEvent(
			WsEvent_SendMessageResponse,
			e.event.Id,
			*comm.NewApiResponse("DB error", comm.CodeInvalidArgs),
		)

		return
	}

	e.origin.responses <- commNewWsEventJSON(
		WsEvent_SendMessageResponse,
		e.event.Id,
		*comm.NewApiResponseWithJson("Success", comm.CodeSuccess, msg),
	)

	targetClients := hub.getClients(msg.ToId)
	if targetClients == nil {
		return
	}

	for _, cl := range targetClients {
		cl.responses <- commNewWsEventJSON(
			WsEvent_NewMessageEvent,
			e.event.Id,
			*comm.NewApiResponseWithJson("New message", comm.CodeSuccess, msg),
		)
	}
}
