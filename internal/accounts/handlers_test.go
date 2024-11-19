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
	"testing"
	"unsafe"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
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
	accs := getPckgTestingAccountsInfo()

	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock()
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
			if resp.StatusCode != test.expectedStatus {

				t.Errorf("Wrong status, expected %d, got %d", test.expectedStatus, resp.StatusCode)
			}
			success := resp.StatusCode == 200

			respBody, _ := io.ReadAll(resp.Body)

			if success {
				respJson := comm.ApiResponseWith[UsernameData]{}

				err := json.Unmarshal(respBody, &respJson)
				if err != nil {
					t.Error("Invalid response format")
				}

				if respJson.Result.Code != test.expectedCommStatusCode {
					t.Errorf("Invalid response comm status code, expected %d, got %d", test.expectedCommStatusCode, respJson.Result.Code)
				}

				username := respJson.Result.Obj
				if username.UserID != test.expectedUid {
					t.Errorf("Got wrong uid, expected %s, got %s", test.expectedUid, username.UserID)
				}
				if username.Id != test.username {
					t.Errorf("Got wrog username, expected %s, got %s", test.username, username.Id)
				}
			} else {
				respJson := comm.ApiResponsePlain{}

				err := json.Unmarshal(respBody, &respJson)
				if err != nil {
					t.Error("Invalid response format")
				}

				if respJson.Result.Code != test.expectedCommStatusCode {
					t.Errorf("Invalid response comm status code, expected %d, got %d", test.expectedCommStatusCode, respJson.Result.Code)
				}
			}
		})
	}
}

func Test_handleGetUser(t *testing.T) {
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
	authMock := setupPckgAuthMock()
	routes := getRoutes(authMock, mongoInst)

	for _, test := range tests {
		reqUrl := fmt.Sprintf("/users/id/%s", test.uid)
		req, _ := http.NewRequest("GET", reqUrl, nil)
		req.Header.Set("Authorization", "Bearer "+test.authToken)
		rec := httptest.NewRecorder()

		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		if resp.StatusCode != test.expectedStatus {
			t.Errorf("Wrong status, expected %d, got %d", test.expectedStatus, resp.StatusCode)
		}
		success := resp.StatusCode == 200

		respBody, _ := io.ReadAll(resp.Body)

		if success {
			respJson := comm.ApiResponseWith[UserData]{}
			err := json.Unmarshal(respBody, &respJson)
			if err != nil {
				t.Error("Invalid response format")
			}
			if respJson.Result.Code != test.expectedCommStatusCode {
				t.Error("Invalid response comm status code")
			}

			userdata := respJson.Result.Obj
			if userdata.Id != test.uid {
				t.Errorf("Got wrong uid, expected %s, got %s", test.uid, userdata.Id)
			}
			if userdata.Email != test.expectedEmail {
				t.Errorf("Got wrong email, expected %s, got %s", test.expectedEmail, userdata.Email)
			}
			if userdata.Username != test.expectedUsername {
				t.Errorf("Got wrong username, expected %s, got %s", test.expectedUsername, userdata.Email)
			}
		} else {
			respJson := comm.ApiResponsePlain{}

			err := json.Unmarshal(respBody, &respJson)
			if err != nil {
				t.Error("Invalid response format")
			}

			if respJson.Result.Code != test.expectedCommStatusCode {
				t.Errorf("Invalid response comm status code, expected %d, got %d", test.expectedCommStatusCode, respJson.Result.Code)
			}
		}
	}
}

func Test_handleRegister(t *testing.T) {
	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock()
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

		if resp.StatusCode != test.status {
			t.Errorf("Invalid http status, expected %d, got %d", test.status, resp.StatusCode)
		}

		respBody, _ := io.ReadAll(resp.Body)
		t.Log(string(respBody))

		respJSON := comm.ApiResponsePlain{}
		err := json.Unmarshal(respBody, &respJSON)
		if err != nil {
			t.Error("Invalid response format")
		}

		if test.commStatusCode != respJSON.Result.Code {
			t.Errorf("Invalid comm response code, expected %d, got %d", test.commStatusCode, respJSON.Result.Code)
		}

		if respJSON.Result.Code == comm.CodeSuccess {
			ctx := context.Background()
			// see if database has correct records
			// TODO: use DBUserRegisterCompleted ??
			if DBGetUserDataUtil(ctx, mongoInst, test.mockUid, nil) != UtilErrorOk {
				t.Error("Record in db is not present")
			}
			if DBGetUsernameDataUtil(ctx, mongoInst, test.username, nil) != UtilErrorOk {
				t.Error("Record in db is not present")
			}

		}
	}
}

func Test_handleUpdateAvatar(t *testing.T) {
	mongoInst, _ := database.NewTestMongoDBInstance()
	authMock := setupPckgAuthMock()
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
		if err != nil {
			t.Error(testutils.InvalidCommResponseFormatMessage)
		}
		if resp.StatusCode != test.expectedStatus {
			t.Error(testutils.InvalidResponseHttpStatusCodeMessage(test.expectedCommStatusCode, resp.StatusCode)) //"Invalid resplonse status code")
		}
		if respJson.Result.Code != test.expectedCommStatusCode {
			t.Error(testutils.InvalidResponseCommStatusCodeMessage(test.expectedCommStatusCode, respJson.Result.Code))
		}

	}
}
