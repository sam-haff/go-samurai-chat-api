package accounts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unsafe"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/testutils"
)

func getRoutes(mockAuth *auth.MockFbAuth, mongoInst *database.MongoDBInstance) *gin.Engine {
	authMock := mockAuth
	testMongoInst := mongoInst

	routes := gin.Default()
	routes.Use(database.InjectDB(testMongoInst), auth.InjectAuth(authMock))
	authRoutes := routes.Group("/", auth.AuthMiddleware)
	publicRoutes := routes.Group("/")
	RegisterHandlers(authRoutes, publicRoutes)

	return routes
}
func Test_handeGetUid(t *testing.T) {
	assert := assert.New(t)

	accs := getPckgTestingAccountsInfo()

	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock(true)
	routes := getRoutes(authMock, mongoInst)

	tests := []struct {
		name                   string
		authToken              string
		username               string
		expectedUid            string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"For existing user", accs[0].Token, accs[0].Username, accs[0].Uid, http.StatusOK, comm.CodeSuccess},
		{"For non-existing user", accs[0].Token, "kkk", "", http.StatusBadRequest, comm.CodeUserNotRegistered},
		{"Unauthorized", "bpmbpm", accs[0].Username, accs[0].Uid, http.StatusUnauthorized, comm.CodeNotAuthenticated},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqUrl := fmt.Sprintf("/uid/%s", test.username)
			req, _ := http.NewRequest("GET", reqUrl, nil)
			req.Header.Set("Authorization", "Bearer "+test.authToken)
			rec := httptest.NewRecorder()

			routes.ServeHTTP(rec, req)

			resp := rec.Result()
			assert.Equal(test.expectedStatus, resp.StatusCode)

			success := resp.StatusCode == 200

			respBody, _ := io.ReadAll(resp.Body)

			if success {
				respJson := comm.ApiResponseWith[UsernameData]{}

				err := json.Unmarshal(respBody, &respJson)
				assert.Nil(err, "invalid response format")
				assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "invalid response comm status code")

				username := respJson.Result.Obj
				assert.Equal(test.expectedUid, username.UserID, "wrong uid")
				assert.Equal(test.username, username.Id)
			} else {
				respJson := comm.ApiResponsePlain{}

				err := json.Unmarshal(respBody, &respJson)
				assert.Nil(err, "invalid response format")
				assert.Equal(test.expectedCommStatusCode, respJson.Result.Code)
			}
		})
	}
}

func Test_handleGetUser(t *testing.T) {
	assert := assert.New(t)

	accs := getPckgTestingAccountsInfo()

	//TODO: add test names
	tests := []struct {
		authToken              string
		uid                    string
		expectedUsername       string
		expectedEmail          string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{accs[0].Token, accs[0].Uid, accs[0].Username, accs[0].Email, http.StatusOK, comm.CodeSuccess},
		{accs[0].Token, "rrr", "", "", http.StatusBadRequest, comm.CodeUserNotRegistered},
		{"bpmbpm", accs[0].Uid, accs[0].Username, accs[0].Email, http.StatusUnauthorized, comm.CodeNotAuthenticated},
	}

	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock(true)
	routes := getRoutes(authMock, mongoInst)

	for _, test := range tests {
		reqUrl := fmt.Sprintf("/users/id/%s", test.uid)
		req, _ := http.NewRequest("GET", reqUrl, nil)
		req.Header.Set("Authorization", "Bearer "+test.authToken)
		rec := httptest.NewRecorder()

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		success := resp.StatusCode == 200
		respBody, _ := io.ReadAll(resp.Body)

		assert.Equal(test.expectedStatus, resp.StatusCode, "wrong status code")
		if success {
			respJson := comm.ApiResponseWith[UserData]{}
			err := json.Unmarshal(respBody, &respJson)
			assert.Nil(err, "invalid response format")
			assert.Equal(test.expectedCommStatusCode, respJson.Result.Code)

			userdata := respJson.Result.Obj
			assert.Equal(test.uid, userdata.Id, "wrong uid")
			assert.Equal(test.expectedEmail, userdata.Email, "wrong email")
			assert.Equal(test.expectedUsername, userdata.Username, "wrong username")
		} else {
			respJson := comm.ApiResponsePlain{}

			err := json.Unmarshal(respBody, &respJson)
			assert.Nil(err, "invalid response format")
			assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "wrong response comm status code")
		}
	}
}

func Test_handleRegister(t *testing.T) {
	assert := assert.New(t)

	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock(true)
	accs := getPckgTestingAccountsInfo()

	routes := getRoutes(authMock, mongoInst)

	// TODO test invalid binding(missing field)
	tests := []struct {
		name           string
		username       string
		email          string
		password       string
		mockUid        string
		status         int
		commStatusCode int
		mockSetup      bool
		mockErr        error
	}{
		{"New user", "registertest1", "registertest1@t.com", "passwo", "registertestuid1", http.StatusOK, comm.CodeSuccess, true, nil},
		{"New user, invalid password format", "registertest2", "registertest2@t.com", "pa", "registertestuid2", http.StatusBadRequest, comm.CodeInvalidArgs, true, nil},
		{"New user, empty password", "registertest2", "registertest2@t.com", "", "registertestuid2", http.StatusBadRequest, comm.CodeInvalidArgs, true, nil},
		{"New user, invalid username format", "lal", "lala2@t.com", "passwo", "registertestuid3", http.StatusBadRequest, comm.CodeInvalidArgs, true, nil},
		{"Existing user 1", "registertest1", "registertest1@t.com", "passwo", "registertestuid1", http.StatusBadRequest, comm.CodeUsernameTaken, true, errors.New("Error placeholder")},
		{"Existing user 2", accs[0].Username, accs[0].Email, "passwo", accs[0].Uid, http.StatusBadRequest, comm.CodeUsernameTaken, true, errors.New("Error placeholder")},
		{"Existing user 3, same email, different username", "registertest2", accs[0].Email, "passwo", "registertestuid3", http.StatusBadRequest, comm.CodeCantCreateAuthUser, true, errors.New("Error placeholder")},
	}

	for _, test := range tests {
		if test.mockSetup {
			r := &fbauth.UserRecord{UserInfo: &fbauth.UserInfo{}}
			r.Email = test.email
			r.UID = test.mockUid
			r.UserInfo.UID = test.mockUid

			matcher := mock.MatchedBy(func(user *fbauth.UserToCreate) bool {
				// using unsafe, because don't want to make unnecessary allocations by
				// creating seperate intermediate type just to enable testing
				mPtr := unsafe.Pointer(user)
				m := *(*map[string]interface{})(mPtr)
				return m["email"] == test.email && m["password"] == test.password
			})

			authMock.On("CreateUser", mock.Anything, matcher).Return(r, test.mockErr).Once()
			authMock.On("CreateUser", mock.Anything, matcher).Return(nil, errors.New("Error placeholder")).Once() // mb dont need once there
		}
	}
	authMock.On("CreateUser", mock.Anything).Return(nil, errors.New("Error placeholder"))

	for _, test := range tests {
		//cant use t.Run here because t.Run runs goroutine but tests are not stateless
		t.Log("=== RUN Test_handleRegister/" + test.name)

		reqParams := RegisterParams{
			Username: test.username,
			Pwd:      test.password,
			Email:    test.email,
		}
		reqBody, _ := json.Marshal(&reqParams)
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
		rec := httptest.NewRecorder()

		req.Header.Set("Content-Type", "application/json")

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		assert.Equal(test.status, resp.StatusCode, "wrong http status")

		respBody, _ := io.ReadAll(resp.Body)

		respJSON := comm.ApiResponsePlain{}
		err := json.Unmarshal(respBody, &respJSON)
		assert.Nil(err, "invalid response format")
		assert.Equal(test.commStatusCode, respJSON.Result.Code)

		if respJSON.Result.Code == comm.CodeSuccess {
			ctx := context.Background()
			// see if database has correct records
			// TODO: use DBUserRegisterCompleted ??
			assert.Equal(UtilErrorOk, DBGetUserDataUtil(ctx, mongoInst, test.mockUid, nil), "no db record")
			assert.Equal(UtilErrorOk, DBGetUsernameDataUtil(ctx, mongoInst, test.username, nil), "no db record")
		}
	}
}

func Test_handleUpdateAvatar(t *testing.T) {
	assert := assert.New(t)

	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock(true)
	routes := getRoutes(authMock, mongoInst)

	accs := getPckgTestingAccountsInfo()

	tests := []struct {
		name                   string
		authAccToken           string
		authAccUid             string
		url                    string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"Good url", accs[0].Token, accs[0].Uid, "http://example.com/exa.jpg", http.StatusOK, comm.CodeSuccess},
		{"Empty url", accs[0].Token, accs[0].Uid, "", http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Not url", accs[0].Token, accs[0].Uid, "abcde/klmnp", http.StatusBadRequest, comm.CodeInvalidArgs},
		{"No auth", "invalid", "invalid", "http://example.com/exa.jpg", http.StatusUnauthorized, comm.CodeNotAuthenticated},
	}

	for _, test := range tests {
		params := UpdateAvatarParams{test.url}
		paramsJsonBytes, _ := json.Marshal(&params)
		req, _ := http.NewRequest("POST", "/updateavatar", bytes.NewBuffer(paramsJsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.authAccToken))

		rec := httptest.NewRecorder()

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		respBytes, _ := io.ReadAll(resp.Body)
		respJson := comm.ApiResponsePlain{}
		err := json.Unmarshal(respBytes, &respJson)

		assert.Nil(err, "invalid response format")
		assert.Equal(test.expectedStatus, resp.StatusCode, "wrong http status")
		assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "wrong comm status")
	}
}

func Test_handleCompleteRegister(t *testing.T) {
	assert := assert.New(t)

	authMock := setupPckgAuthMock(false)
	accsNotInDB := auth.GetTestingAccountsInfo(pckgPrefix, TestingAccountsInDBCount+1, 2)
	for _, acc := range accsNotInDB {
		authMock.AddMockTestingAccount(acc)
	}
	auth.FinalizeSetupAuthMock(authMock)

	mongoInst, _ := database.NewTestMongoDBInstance()
	routes := getRoutes(authMock, mongoInst)
	accs := getPckgTestingAccountsInfo()

	tests := []struct {
		name                   string
		username               string
		token                  string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"Incomplete registered auth", accsNotInDB[0].Username, accsNotInDB[0].Token, http.StatusOK, comm.CodeSuccess},
		{"Already registered auth", accs[0].Username, accs[0].Token, http.StatusBadRequest, comm.CodeCantCreateAuthUser},
		{"Invalid username format 1", "abc", accsNotInDB[1].Token, http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Invalid username format 2", "abcd_d", accsNotInDB[1].Token, http.StatusBadRequest, comm.CodeInvalidArgs},
	}

	for _, test := range tests {
		t.Log("=== RUN Test_handleCompleteRegister/" + test.name)

		params := CompleteRegisterParams{Username: test.username}
		paramsBytes, _ := json.Marshal(params)
		req, _ := http.NewRequest("POST", "/completeregister", bytes.NewBuffer(paramsBytes))
		//req.Header.Set("Authorization", "Bearer "+test.token)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", test.token))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		assert.Equal(test.expectedStatus, resp.StatusCode, "wrong http status code")
		respJson := comm.ApiResponsePlain{}
		respJsonBytes, _ := io.ReadAll(resp.Body)
		err := json.Unmarshal(respJsonBytes, &respJson)
		assert.Nil(err, "invalid comm response format")
		assert.Equal(test.expectedCommStatusCode, respJson.Result.Code, "wrong comm status code")
		t.Log(respJson.Result.Msg)
	}
}

func Test_handleRegisterToken(t *testing.T) {
	assert := assert.New(t)

	authMock := setupPckgAuthMock(false)
	accsNotInDB := auth.GetTestingAccountsInfo(pckgPrefix, TestingAccountsInDBCount+10, 2)
	for _, acc := range accsNotInDB {
		authMock.AddMockTestingAccount(acc)
	}
	auth.FinalizeSetupAuthMock(authMock)

	mongoInst, _ := database.NewTestMongoDBInstance()
	accs := getPckgTestingAccountsInfo()
	routes := getRoutes(authMock, mongoInst)

	tooLongDeviceName := strings.Repeat("a", MaxFcmDeviceNameLength+1)
	tooLongToken := strings.Repeat("b", MaxFcmTokenLength+1)

	tests := []struct {
		name                   string
		token                  string
		uid                    string
		fcmToken               string
		fcmDeviceName          string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"Normal", accs[0].Token, accs[0].Uid, "ccc", "dddd", http.StatusOK, comm.CodeSuccess},
		{"Too long token", accs[0].Token, accs[0].Uid, tooLongToken, "dddd", http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Too long device name", accs[0].Token, accs[0].Uid, "eee", tooLongDeviceName, http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Empty token", accs[0].Token, accs[0].Uid, "", "dddd", http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Empty device name", accs[0].Token, accs[0].Uid, "eee", "", http.StatusBadRequest, comm.CodeInvalidArgs},
		{"Not fully registered", accsNotInDB[0].Token, accs[0].Uid, "eee", "ddd", http.StatusUnauthorized, comm.CodeUserNotRegistered},
	}

	for _, test := range tests {
		testutils.PrintTestName(t, test.name)

		params := RegisterTokenParams{Token: test.fcmToken, DeviceName: test.fcmDeviceName}
		paramsBytes, _ := json.Marshal(&params)
		req, _ := http.NewRequest("POST", "/registertoken", bytes.NewBuffer(paramsBytes))
		req.Header.Set("Authorization", "Bearer "+test.token)
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
			// token is meant to be registered succesfully, need to check if it's actually true
			userData := UserData{}
			utilStatus := DBGetUserDataUtil(context.Background(), mongoInst, test.uid, &userData)
			assert.Equal(UtilErrorOk, utilStatus, "failed to get user record")
			dbFcmToken, ok := userData.Tokens[test.fcmDeviceName]
			assert.True(ok, "registered token is not present")
			assert.Equal(test.fcmToken, dbFcmToken, "wrong value of registered token")
		}
	}

}
