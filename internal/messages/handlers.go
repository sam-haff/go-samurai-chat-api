package messages

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"
	"go-chat-app-api/internal/users"
)

func RegisterHandlers(routers *gin.Engine) {
	routers.POST("/addmessage", middleware.AuthMiddleware, handleAddMessage)
	routers.POST("/chat", middleware.AuthMiddleware, handleGetChat)
}

type AddMessageParams struct {
	Msg  string `json:"text" binding:"gte=1,lte=4096,required"`
	ToId string `json:"to" binding:"gte=1,lte=1024,required"`
}

type Ids []string

type MessageDataWithId struct {
	Id  string      `bson:"msg_id" json:"msg_id"`
	Msg MessageData `bson:"msg" json:"msg"`
}

func composeChatKey(uid1 string, uid2 string) string {
	ids := []string{uid1, uid2}
	sort.Strings(ids)

	compIndex := ids[0] + ids[1]
	return compIndex
}

func fcmSendNewMessage(ctx *gin.Context, tokens map[string]string, msg MessageData, needsNotification bool, needsMsg bool) bool {

	fbApp := ctx.MustGet(middleware.CtxVarFirebaseApp).(*firebase.App)
	fcmClient, _ := fbApp.Messaging(ctx)

	isNotification := 0
	if needsNotification {
		isNotification = 1
	}
	isMsg := 0
	if needsMsg {
		isMsg = 1
	}

	fcmMsgData := map[string]string{
		"is_notification": strconv.FormatInt(int64(isNotification), 10),
		"is_msg":          strconv.FormatInt(int64(isMsg), 10),
	}

	if needsNotification {
		fcmMsgData["click_action"] = "FLUTTER_NOTIFICATION_CLICK"
	}
	if needsMsg {
		fcmMsgData["_from"] = msg.FromId
		fcmMsgData["to"] = msg.ToId
		fcmMsgData["username"] = msg.FromUsername
		fcmMsgData["msg"] = msg.Text
		fcmMsgData["img_url"] = msg.ImgUrl
		fcmMsgData["created_at"] = strconv.FormatInt(msg.CreatedAt, 10)
	}

	for _, token := range tokens {
		fcmMsg := &messaging.Message{}

		fcmMsg.Token = token
		fcmMsg.Data = fcmMsgData
		if needsNotification {
			fcmMsg.Notification = &messaging.Notification{
				Title: msg.FromUsername,
				Body:  msg.Text,
			}
			fcmMsg.Android = &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Sound: "default",
				},
			}
		}

		_, err := fcmClient.Send(
			ctx,
			fcmMsg,
		)

		if err != nil {
			fmt.Printf("Failed to send FCM message %s \n", err.Error())
			// some of tokens can be invalid
			// mb submit them for cleaning???

			//ctx.String(400, apiResponse("Failed to send FCM messages", CodeInvalidArgs))
			//return false
		}

	}
	return true
}
func handleAddMessage(ctx *gin.Context) {
	userId := ctx.MustGet(middleware.CtxVarUserId).(string)
	if len(userId) == 0 {
		comm.AbortUnauthorized(ctx, "Invalid creds", comm.CodeNotAuthenticated)
		return
	}
	params := AddMessageParams{}
	if err := ctx.ShouldBind(&params); err != nil {
		comm.AbortFailedBinding(ctx, err)
		return
	}

	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)

	fromUserData := users.UserData{}
	if !users.GetUserData(ctx, userId, &fromUserData) {
		return
	}
	toUserData := users.UserData{}
	if !users.GetUserData(ctx, params.ToId, &toUserData) {
		return
	}

	compIndex := composeChatKey(userId, params.ToId)

	msg := MessageData{
		ConversationID: compIndex,
		Text:           params.Msg,
		FromId:         userId,
		ToId:           params.ToId,
		FromUsername:   fromUserData.Username,
		ImgUrl:         fromUserData.Img_url,
		CreatedAt:      time.Now().UnixMilli(),
	}

	messagesCollection := mongoInst.Collection(database.MessagesCollection)

	_, err := messagesCollection.InsertOne(ctx, msg)

	if err != nil {
		respMsg := fmt.Sprintf("Failed to write messages to db with: %s", err.Error())
		comm.AbortBadRequest(ctx, respMsg, comm.CodeInvalidArgs)
		return
	}

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
	comm.GenericOK(ctx)
}

type GetChatParams struct {
	Limit           int    `json:"limit"`
	BeforeTimeStamp int64  `json:"before_timestamp"`
	With            string `json:"with" binding:"lte=1024,required"`
	Inverse         bool   `json:"inverse"`
}

func handleGetChat(ctx *gin.Context) {
	userId := ctx.MustGet(middleware.CtxVarUserId).(string)
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

	mongoInst := ctx.MustGet(middleware.CtxVarMongoDBInst).(*database.MongoDBInstance)
	messagesCollection := mongoInst.Collection(database.MessagesCollection)

	// TODO dont allow high limit to mitigate possible attack
	opts := options.Find().SetLimit(int64(params.Limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})

	compKey := composeChatKey(userId, params.With)

	fmt.Printf("Gettting chat with comp key %s\n", compKey)

	filter := bson.D{
		{
			Key: "$and", Value: bson.A{
				bson.D{{Key: "conv_id", Value: compKey}},
				bson.D{{Key: "created_at", Value: bson.D{{Key: "$lt", Value: params.BeforeTimeStamp}}}},
			},
		}}
	cursor, err := messagesCollection.Find(ctx, filter, opts)
	if err != nil {
		comm.AbortBadRequest(ctx, "Failed to fetch messages", comm.CodeInvalidArgs)
		return
	}

	var messages []MessageData

	err = cursor.All(ctx, &messages)
	if err != nil {
		comm.AbortBadRequest(ctx, "Couldnt parse data from db", comm.CodeInvalidArgs)
		return
	}

	comm.GenericOKJSON(ctx, messages)
}
