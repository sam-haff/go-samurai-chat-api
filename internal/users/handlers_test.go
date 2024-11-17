package users

import (
	"encoding/json"
	"fmt"
	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_handleGetUser(t *testing.T) {
	accs := GetTestingAccountsInfo()
	authMock := SetupAuthMock()
	testMongoInst, _ := database.NewTestMongoDBInstance()

	tests := []struct {
		authToken              string
		uid                    string
		expectedUsername       string
		expectedEmail          string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{accs[0].token, accs[0].uid, accs[0].username, accs[0].email, http.StatusOK, comm.CodeSuccess},
		{accs[0].token, "rrr", "", "", http.StatusBadRequest, comm.CodeUserNotRegistered},
		{"bpmbpm", accs[0].uid, accs[0].username, accs[0].email, http.StatusUnauthorized, comm.CodeNotAuthenticated},
	}

	routers := gin.Default()
	routers.GET("/:uid", middleware.InjectParams(nil, authMock, testMongoInst), middleware.AuthMiddleware, handleGetUser)

	for _, test := range tests {
		t.Log("Test: " + test.authToken)
		reqUrl := fmt.Sprintf("/%s", test.uid)
		req, _ := http.NewRequest("GET", reqUrl, nil)
		req.Header.Set("Authorization", "Bearer "+test.authToken)
		rec := httptest.NewRecorder()

		routers.ServeHTTP(rec, req)

		resp := rec.Result()
		if resp.StatusCode != test.expectedStatus {
			t.Errorf("Wrong status, expected %d, got %d", test.expectedStatus, resp.StatusCode)
		}
		success := resp.StatusCode == 200

		respBody, _ := io.ReadAll(resp.Body)

		if success {
			respJson := comm.ApiResponseWith[accounts.UserData]{}
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
