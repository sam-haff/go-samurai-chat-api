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
	authRoutes.POST("/addmessage", accounts.CompleteRegisteredMiddleware, handleAddMessage)
	authRoutes.POST("/chat", accounts.CompleteRegisteredMiddleware, handleGetChat)
}

const (
	MaxMessageLength = 4096
	MaxIdLength      = 1024
)

type AddMessageParams struct {
	Msg  string `json:"text" binding:"min=1,max=4096,required"`
	ToId string `json:"to" binding:"min=1,max=1024,required"`
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
	params := AddMessageParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)

	// TODO: set userdata in CompleteRegisteredMiddleware to avoid duplicate requests
	fromUserData := ctx.MustGet(accounts.CtxVarUserData).(accounts.UserData)
	toUserData := accounts.UserData{}
	if !accounts.DBGetUserData(ctx, params.ToId, &toUserData) {
		return
	}

	msg := NewMessageData(fromUserData, params.ToId, params.Msg)
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

func handleGetChat(ctx *gin.Context) {
	userId := ctx.MustGet(auth.CtxVarUserId).(string) // 500 if no auth middleware
	if len(userId) == 0 {
		return
	}

	params := GetChatParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		fmt.Printf("Invalid params\n")
		comm.AbortFailedBinding(ctx, err)
		return
	}

	fmt.Printf("Getting msgs before %d\n", params.BeforeTimeStamp)

	mongoInst := ctx.MustGet(database.CtxVarMongoDBInst).(*database.MongoDBInstance)
	var messages []MessageData
	res := DBGetMessagesUtil(ctx, mongoInst, userId, params.With, params.Limit, false, params.BeforeTimeStamp, &messages)
	if res == UtilStatusNotFound {
		comm.AbortBadRequest(ctx, "Failed to fetch messages", comm.CodeInvalidArgs)
		return
	}
	if res == UtilStatusCantParse {
		comm.AbortBadRequest(ctx, "Couldnt parse data from db", comm.CodeInvalidArgs)
		return
	}

	comm.GenericOKJSON(ctx, messages)
}
