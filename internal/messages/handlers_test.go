package messages

import (
	"bytes"
	"context"
	"encoding/json"
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/testutils"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getRoutes(mockAuth *auth.MockFbAuth, mongoInst *database.MongoDBInstance, fcm *FcmClientMock) *gin.Engine {
	authMock := mockAuth
	testMongoInst := mongoInst
	fcmMock := fcm

	routes := gin.Default()
	routes.Use(database.InjectDB(testMongoInst), auth.InjectAuth(authMock), InjectFcm(fcmMock))
	authRoutes := routes.Group("/", auth.AuthMiddleware)
	publicRoutes := routes.Group("/")
	RegisterHandlers(authRoutes, publicRoutes)

	return routes
}

func getFcmMsgMatcher(fcmToken string, title string, body string) func(*messaging.Message) bool {
	return func(msg *messaging.Message) bool {
		return msg.Token == fcmToken && msg.Notification.Title == title && msg.Notification.Body == body
	}
}
func Test_handleAddMessage(t *testing.T) {
	assert := assert.New(t)

	accs := getPckgTestingAccountsInfo()
	authMock := setupPckgAuthMock(false)
	mongoInst, _ := database.NewTestMongoDBInstance()
	fcmMock := &FcmClientMock{}
	routes := getRoutes(authMock, mongoInst, fcmMock)

	accNotInDB := accounts.GetTestingAccountInfo(pckgPrefix, TestingAccountsInDBCount)
	authMock.AddMockTestingAccount(accNotInDB.ToTestingAuthRecord())
	auth.FinalizeSetupAuthMock(authMock)

	msgTooLarge := strings.Repeat("a", MaxMessageLength+1)
	accIdTooLarge := accounts.GetTestingAccountInfo(pckgPrefix, TestingAccountsInDBCount) //TODO remove duplciate call
	accIdTooLarge.Id = strings.Repeat("a", MaxIdLength+1)

	compKey := composeChatKey(accs[0].Id, accs[1].Id)
	msg1 := MessageData{
		FromId:         accs[0].Id,
		ToId:           accs[1].Id,
		FromUsername:   accs[0].Username,
		Text:           "Message content 1",
		ConversationID: compKey}
	msg2 := MessageData{
		FromId:         accs[1].Id,
		ToId:           accs[0].Id,
		FromUsername:   accs[1].Username,
		Text:           "Message content 2",
		ConversationID: compKey}
	//tests
	//cases: normal, incomplete registration, invalid receiver id, invalid binding
	// + check final db records to ensure that msgs are added correcctly
	tests := []struct {
		name                   string
		sender                 accounts.TestingAccount
		receiver               accounts.TestingAccount
		msg                    string
		expectedResultingChat  []MessageData
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"Normal, first message", accs[0], accs[1], msg1.Text, []MessageData{msg1}, http.StatusOK, comm.CodeSuccess},
		{"Normal, response message", accs[1], accs[0], msg2.Text, []MessageData{msg2, msg1}, http.StatusOK, comm.CodeSuccess},
		{"Receiver not fully registered", accs[0], accNotInDB, "Message content", []MessageData{}, http.StatusBadRequest, comm.CodeUserNotRegistered},
		{"Invalid msg format", accs[0], accs[1], msgTooLarge, []MessageData{}, http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Invalid id format", accs[0], accIdTooLarge, "Message content", []MessageData{}, http.StatusBadRequest, comm.CodeInvalidArgs},
	}

	//setup fcm mock
	for _, test := range tests {
		if test.expectedStatus == http.StatusOK {
			fcmMock.On("Send", mock.Anything, mock.MatchedBy(getFcmMsgMatcher(test.receiver.Tokens["device"], test.sender.Username, test.msg))).Return("", nil)
		}
	}
	fcmMock.On("Send", mock.Anything, mock.Anything).Return("", nil) // TODO make it more detailed to check correctness of the FCM message submit

	for _, test := range tests {
		testutils.PrintTestName(t, test.name)
		params := AddMessageParams{ToId: test.receiver.Id, Msg: test.msg}
		paramsBytes, _ := json.Marshal(&params)
		req, _ := http.NewRequest("POST", "/addmessage", bytes.NewBuffer(paramsBytes))
		req.Header.Set("Authorization", "Bearer "+test.sender.Token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		assert.Equal(test.expectedStatus, resp.StatusCode, "wrong http status code")
		respJson := comm.ApiResponsePlain{}
		respJsonBytes, _ := io.ReadAll(resp.Body)
		err := json.Unmarshal(respJsonBytes, &respJson)
		assert.Nil(err, "invalid response format")
		assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "wrong comm status code")

		//check correctness of db records
		if resp.StatusCode == http.StatusOK {
			var dbMsgs []MessageData
			res := DBGetMessagesUtil(context.Background(), mongoInst, test.sender.Id, test.receiver.Id, 1024, false, math.MaxInt64, &dbMsgs)
			assert.Equal(UtilStatusOk, res, "failed to get messages from db")
			assert.Equal(len(test.expectedResultingChat), len(dbMsgs), "wrong number of messages in db")
			if len(test.expectedResultingChat) == len(dbMsgs) {
				for i, msg := range test.expectedResultingChat {
					assert.Equal(msg.FromId, dbMsgs[i].FromId, "wrong id of sender")
					assert.Equal(msg.ToId, dbMsgs[i].ToId, "wrong id of receiver")
					assert.Equal(msg.Text, dbMsgs[i].Text, "wrong message text")
					assert.Equal(msg.FromUsername, dbMsgs[i].FromUsername, "wrong username of sender")
					assert.Equal(msg.ConversationID, dbMsgs[i].ConversationID, "wrong message conversation id")
				}
			}
		}

	}

	fcmMock.AssertExpectations(t)
}

func Test_handleGetChat(t *testing.T) {
	assert := assert.New(t)

	accs := getPckgTestingAccountsInfo()
	authMock := setupPckgAuthMock(true)
	mongoInst, _ := database.NewTestMongoDBInstance()
	routes := getRoutes(authMock, mongoInst, nil)
	accNotInDB := accounts.GetTestingAccountsInfo(pckgPrefix, TestingAccountsInDBCount+10, 1)[0] //getPckgTestingAccountsInfo()
	accIdTooLarge := accNotInDB
	accIdTooLarge.Id = strings.Repeat("a", 2000)

	ctx := context.Background()
	acc1 := accs[0]
	acc2 := accs[2]
	msgsToAdd := []MessageData{
		NewMessageData(acc1.UserData, acc2.Id, "Hello"),
		NewMessageData(acc2.UserData, acc1.Id, "Hi"),
		NewMessageData(acc1.UserData, acc2.Id, "How are you?"),
	}
	for _, msg := range msgsToAdd {
		DBAddMessageUtil(ctx, mongoInst, msg)
	}

	tests := []struct {
		name                        string
		authAcc                     accounts.TestingAccount
		to                          accounts.TestingAccount
		limit                       int
		expectedMessagesInSendOrder []MessageData
		expectedStatus              int
		expectedCommStatusCode      int
	}{
		{"Normal 1", acc1, acc2, 10, msgsToAdd, http.StatusOK, comm.CodeSuccess},
		{"Normal 2", acc2, acc1, 10, msgsToAdd, http.StatusOK, comm.CodeSuccess},
		{"Invalid receiver", acc1, accNotInDB, 10, msgsToAdd, http.StatusBadRequest, comm.CodeUserNotRegistered},
		{"Empty chat", acc1, accs[3], 10, []MessageData{}, http.StatusOK, comm.CodeSuccess},
		{"Too big limit", acc1, acc2, 10000000, msgsToAdd, http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Too large To id", acc1, accIdTooLarge, 10, msgsToAdd, http.StatusBadRequest, comm.CodeInvalidArgs},
	}

	for _, test := range tests {
		testutils.PrintTestName(t, test.name)

		params := GetChatParams{Limit: test.limit, BeforeTimeStamp: math.MaxInt64, With: test.to.Id, Inverse: false}
		paramsBytes, _ := json.Marshal(params)
		req, _ := http.NewRequest("POST", "/chat", bytes.NewBuffer(paramsBytes))
		rec := httptest.NewRecorder()
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+test.authAcc.Token)

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		assert.Equal(test.expectedStatus, resp.StatusCode, "wrong http status code")
		respJsonBytes, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			respJson := comm.ApiResponseWith[[]MessageData]{}
			err := json.Unmarshal(respJsonBytes, &respJson)
			assert.Nil(err, "invalid response format")
			assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "wrong comm code")

			assert.Equal(len(test.expectedMessagesInSendOrder), len(respJson.Result.Obj), "got wrong number of messages")
			if len(test.expectedMessagesInSendOrder) == len(respJson.Result.Obj) {
				for i, msg := range slices.Backward(test.expectedMessagesInSendOrder) {
					assert.Equal(msg.FromId, respJson.Result.Obj[i].FromId)
					assert.Equal(msg.ToId, respJson.Result.Obj[i].ToId)
					assert.Equal(msg.Text, respJson.Result.Obj[i].Text)
					assert.Equal(msg.FromUsername, respJson.Result.Obj[i].FromUsername)
					assert.Equal(msg.ConversationID, respJson.Result.Obj[i].ConversationID)
				}
			}
		} else {
			respJson := comm.ApiResponsePlain{}
			err := json.Unmarshal(respJsonBytes, &respJson)
			assert.Nil(err, "invalid response format")
			assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "wrong comm code")
		}
	}
}
