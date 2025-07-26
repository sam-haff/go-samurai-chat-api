package messages

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
)

func RegisterHandlers(authRoutes *gin.RouterGroup, publicRoutes *gin.RouterGroup) {
	authRoutes.POST("/chats/:with", accounts.CompleteRegisteredMiddleware, handleAddMessage)
	authRoutes.GET("/chats/:with", accounts.CompleteRegisteredMiddleware, handleGetSingleChat)
	authRoutes.GET("/chats", accounts.CompleteRegisteredMiddleware, handleGetChats)
}

func handleGetChats(ctx *gin.Context) {
	// TODO: should be fast with indexed conv_id but mb use caching?
	userId := ctx.MustGet(auth.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	chats := make([]string, 0)
	if status := DBGetChatsUtil(ctx, mongoInst, userId, &chats); status != UtilStatusOk {
		comm.AbortBadRequest(ctx, "Failed to get chats", comm.CodeInvalidArgs)
		return
	}

	comm.GenericOKJSON(ctx, chats)
}

const (
	MaxMessageLength = 4096
	MaxIdLength      = 256
)

type AddMessageParams struct {
	Msg  string `json:"text" binding:"min=1,max=4096,required"`
	ToId string `json:"to" binding:"min=1,max=256,required"`
}
type AddMessageUriParams struct {
	ToId string `uri:"to" binding:"min=1,max=256,required"`
}
type AddMessageJsonParams struct {
	Msg string `json:"text" binding:"min=1,max=4096,required"`
}

type Ids []string

type MessageDataWithId struct {
	Id  string      `bson:"msg_id" json:"msg_id"`
	Msg MessageData `bson:"msg" json:"msg"`
}

func handleAddMessage(ctx *gin.Context) {
	userId := ctx.MustGet(auth.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return
	}

	uriParams := AddMessageUriParams{}
	if err := ctx.ShouldBind(&uriParams); err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}
	withUid := uriParams.ToId

	jsonParams := AddMessageJsonParams{}
	if err := ctx.ShouldBind(&jsonParams); err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	// TODO: set userdata in CompleteRegisteredMiddleware to avoid duplicate requests
	fromUserData := ctx.MustGet(accounts.CtxVarUserData).(accounts.UserData)
	toUserData := accounts.UserData{}
	if !accounts.DBGetUserData(ctx, withUid, &toUserData) {
		return
	}

	msg := NewMessageData(fromUserData.Id, withUid, jsonParams.Msg)
	err := DBAddMessageUtil(ctx, mongoInst, msg)
	if err != nil {
		respMsg := fmt.Sprintf("Failed to write messages to db with: %s", err.Error())
		comm.AbortBadRequest(ctx, respMsg, comm.CodeInvalidArgs)
		return
	}

	//TODO: rework logic after WebSocket introduction
	if toUserData.Tokens != nil {
		if !fcmSendNewMessage(ctx, toUserData.Tokens, msg, true, false) {
			return
		}
		if !fcmSendNewMessage(ctx, toUserData.Tokens, msg, false, true) {
			return
		}
	}
	if fromUserData.Tokens != nil {
		if !fcmSendNewMessage(ctx, fromUserData.Tokens, msg, false, true) {
			return
		}
	}
	// TODO: mb return message id
	comm.GenericOK(ctx)
}

type GetChatParams struct {
	Limit           int    `json:"limit" binding:"max=1024"`
	BeforeTimeStamp int64  `json:"before_timestamp"`
	With            string `json:"with" binding:"max=1024,required"`
	Inverse         bool   `json:"inverse"`
}

type GetSingleChatParams struct {
	Limit           int   `json:"limit" form:"limit" binding:"max=1024"`
	BeforeTimeStamp int64 `json:"before_timestamp" form:"before_timestamp"`
	Inverse         bool  `json:"inverse" form:"inverse"`
}

func handleGetSingleChat(ctx *gin.Context) {
	userId := ctx.MustGet(auth.CtxVarUserId).(string) // 500 if no auth middleware
	if len(userId) == 0 {
		return
	}

	withUid := ctx.Param("with") // TODO: make struct with validation tags
	params := GetSingleChatParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		fmt.Printf("Invalid params\n")
		comm.AbortFailedBinding(ctx, err)
		return
	}

	toUserData := accounts.UserData{}
	if !accounts.DBGetUserData(ctx, withUid, &toUserData) { // correspondent should have valid registration
		return
	}

	fmt.Printf("Getting msgs before %d\n", params.BeforeTimeStamp)

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	var messages []MessageData
	res := DBGetMessagesUtil(ctx, mongoInst, userId, withUid, params.Limit, false, params.BeforeTimeStamp, &messages)
	if res == UtilStatusNotFound {
		comm.AbortBadRequest(ctx, "Failed to fetch messages", comm.CodeInvalidArgs)
		return
	}
	if res == UtilStatusCantParse {
		comm.AbortBadRequest(ctx, "Couldnt parse data from db", comm.CodeInvalidArgs)
		return
	}

	if messages == nil {
		messages = make([]MessageData, 0)
	}

	comm.GenericOKJSON(ctx, messages)
}
