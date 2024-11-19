package accounts

import (
	"encoding/json"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

func Test_CompleteRegisteredMiddleware(t *testing.T) {

	accs := getPckgTestingAccountsInfo()
	authMock := setupPckgAuthMock(false)
	mongoInst, _ := database.NewTestMongoDBInstance()

	testingAccountsNextIndex := TestingAccountsInDBCount
	accNotFullyRegistered := auth.GetTestingAccountInfo(pckgPrefix, testingAccountsNextIndex)
	authToken := &fbauth.Token{}
	authToken.UID = accNotFullyRegistered.Uid
	authToken.Firebase.SignInProvider = "email"
	authMock.On("VerifyToken", mock.Anything, accNotFullyRegistered.Token).Return(authToken, nil)
	auth.FinalizeSetupAuthMock(authMock)

	//add tests wihtout auth middleware
	tests := []struct {
		name                   string
		authToken              string
		uid                    string
		username               string
		email                  string
		expectedStatus         int
		expectedCommStatusCode int
	}{
		{"Auth with completed register", accs[0].Token, accs[0].Uid, accs[0].Username, accs[0].Email, http.StatusOK, comm.CodeSuccess},
		{"Auth with completed register, without auth middleware", accs[0].Token, accs[0].Uid, accs[0].Username, accs[0].Email, http.StatusUnauthorized, comm.CodeNotAuthenticated},
		{"Auth with incomplete register", accNotFullyRegistered.Token, accNotFullyRegistered.Uid, accNotFullyRegistered.Username, accNotFullyRegistered.Email, http.StatusUnauthorized, comm.CodeUserNotRegistered},
	}

	for _, test := range tests {
		routes := gin.Default()
		routes.Use(auth.InjectAuth(authMock), database.InjectDB(mongoInst))
		testHandler := func(ctx *gin.Context) {
			username := ctx.MustGet(CtxVarUserUsername)

			if username != test.username {
				t.Errorf("Wrong username, expected %s, got %s", test.username, username)
			}

			email := ctx.MustGet(CtxVarUserEmail)
			if email != test.email {
				t.Errorf("Wrong username, expected %s, got %s", test.email, email)
			}

			comm.GenericOK(ctx)
		}
		routes.POST("/test", auth.AuthMiddleware, CompleteRegisteredMiddleware, testHandler)

		req, _ := http.NewRequest("POST", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+test.authToken)
		rec := httptest.NewRecorder()
		routes.ServeHTTP(rec, req)

		resp := rec.Result()
		respJsonBytes, _ := io.ReadAll(resp.Body)
		respJson := comm.ApiResponsePlain{}
		err := json.Unmarshal(respJsonBytes, &respJson)

		if resp.StatusCode != test.expectedStatus {

			t.Error(testutils.InvalidResponseHttpStatusCodeMessage(test.expectedStatus, resp.StatusCode))
		}
		if err != nil {
			t.Error(testutils.InvalidCommResponseFormatMessage)
		}
		if respJson.Result.Code != test.expectedCommStatusCode {
			t.Error(testutils.InvalidResponseHttpStatusCodeMessage(test.expectedCommStatusCode, respJson.Result.Code))
		}
	}

}
