package accounts

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
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
	"go-chat-app-api/internal/middleware"
)

func getRoutes(mockAuth *auth.MockFbAuth) *gin.Engine {
	authMock := mockAuth
	testMongoInst, _ := database.NewTestMongoDBInstance()

	routes := gin.Default()
	routes.Use(middleware.InjectParams(nil, authMock, testMongoInst))
	authRoutes := routes.Group("/", middleware.AuthMiddleware)
	publicRoutes := routes.Group("/")
	RegisterHandlers(authRoutes, publicRoutes)

	return routes
}

func Test_handleRegister(t *testing.T) {
	authMock := setupPckgAuthMock()

	accs := getPckgTestingAccountsInfo()
	routes := getRoutes(authMock)

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

	for i, test := range tests {
		log.Printf("%d\n", i)
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
	}
}
