package messages

import (
	"fmt"
	"sort"
	"strconv"

	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
)

func composeChatKey(uid1 string, uid2 string) string {
	ids := []string{uid1, uid2}
	sort.Strings(ids)

	compIndex := ids[0] + ids[1]
	return compIndex
}

func newFcmMessage(token string, msg MessageData, needsNotification bool, needsMsg bool) *messaging.Message {
	fcmMsg := &messaging.Message{}

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
		fcmMsg.APNS = &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true,
				},
			},
		}
	}

	return fcmMsg
}

func fcmSendNewMessage(ctx *gin.Context, tokens map[string]string, msg MessageData, needsNotification bool, needsMsg bool) bool {

	//fbApp := ctx.MustGet(middleware.CtxVarFirebaseApp).(*firebase.App)
	fcmClient, _ := ctx.MustGet(CtxVarFcm).(FcmClient) //fbApp.Messaging(ctx)

	for _, token := range tokens {
		fcmMsg := newFcmMessage(token, msg, needsNotification, needsMsg)

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
