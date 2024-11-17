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

func GetRoutes() *gin.Engine {
	authMock := setupPckgAuthMock()
	testMongoInst, _ := database.NewTestMongoDBInstance()

	routes := gin.Default()
	routes.Use(middleware.InjectParams(nil, authMock, testMongoInst))
	authRoutes := routes.Group("/", middleware.AuthMiddleware)
	publicRoutes := routes.Group("/")
	RegisterHandlers(authRoutes, publicRoutes)

	return routes
}

func Test_handeGetUid(t *testing.T) {
	accs := getPckgTestingAccountsInfo()

	routes := GetRoutes()

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
				respJson := comm.ApiResponseWith[accounts.UsernameData]{}

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

	routes := GetRoutes()

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
