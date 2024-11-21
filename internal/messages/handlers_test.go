package messages

import (
	"bytes"
	"encoding/json"
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func Test_handleAddMessage(t *testing.T) {
	assert := assert.New(t)

	accs := getPckgTestingAccountsInfo()
	authMock := setupPckgAuthMock(false)
	mongoInst, _ := database.NewTestMongoDBInstance()
	fcmMock := &FcmClientMock{}
	fcmMock.On("Send", mock.Anything, mock.Anything) // TODO make it more detailed to check correctness of the FCM message submit
	routes := getRoutes(authMock, mongoInst, fcmMock)

	accNotInDB := accounts.GetTestingAccountInfo(pckgPrefix, TestingAccountsInDBCount)
	authMock.AddMockTestingAccount(accNotInDB.ToTestingAuthRecord())
	auth.FinalizeSetupAuthMock(authMock)

	msgTooLarge := strings.Repeat("a", MaxMessageLength+1)
	accIdTooLarge := accounts.GetTestingAccountInfo(pckgPrefix, TestingAccountsInDBCount) //TODO remove duplciate call
	accIdTooLarge.Id = strings.Repeat("a", MaxIdLength+1)

	//tests
	//cases: normal, incomplete registration, invalid receiver id, invalid binding
	// + check final db records to ensure that msgs are added correcctly
	tests := []struct {
		name                   string
		sender                 accounts.TestingAccount
		receiver               accounts.TestingAccount
		msg                    string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"Normal", accs[0], accs[1], "Message content", http.StatusOK, comm.CodeSuccess},
		{"Receiver not fully registered", accs[0], accNotInDB, "Message content", http.StatusBadRequest, comm.CodeUserNotRegistered},
		{"Invalid msg format", accs[0], accs[1], msgTooLarge, http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Invalid id format", accs[0], accIdTooLarge, "Message content", http.StatusBadRequest, comm.CodeInvalidArgs},
	}

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

		if resp.StatusCode == http.StatusOK {
			// check on correctness of db records
		}
	}

}
