package accounts

import (
	"encoding/json"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/comm"
	"go-chat-app-api/internal/database"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CompleteRegisteredMiddleware(t *testing.T) {
	assert := assert.New(t)

	accs := getPckgTestingAccountsInfo()
	authMock := setupPckgAuthMock(false)
	mongoInst, _ := database.NewTestMongoDBInstance()

	testingAccountsNextIndex := TestingAccountsInDBCount
	accNotFullyRegistered := GetTestingAccountInfo(pckgPrefix, testingAccountsNextIndex)
	authToken := &fbauth.Token{}
	authToken.UID = accNotFullyRegistered.Id
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
		{"Auth with completed register", accs[0].Token, accs[0].Id, accs[0].Username, accs[0].Email, http.StatusOK, comm.CodeSuccess},
		{"Auth with incomplete register", accNotFullyRegistered.Token, accNotFullyRegistered.Id, accNotFullyRegistered.Username, accNotFullyRegistered.Email, http.StatusUnauthorized, comm.CodeUserNotRegistered},
	}

	for _, test := range tests {
		routes := gin.Default()
		routes.Use(auth.InjectAuth(authMock), database.InjectDB(mongoInst))
		testHandler := func(ctx *gin.Context) {
			username := ctx.MustGet(CtxVarUserUsername)
			assert.Equal(test.username, username, "wrong username")

			email := ctx.MustGet(CtxVarUserEmail)
			assert.Equal(test.email, email, "wrong email")

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

		assert.Equal(test.expectedStatus, resp.StatusCode, "wrong http status")
		assert.Nil(err, "wrong resp format")
		assert.Equal(respJson.Result.Code, test.expectedCommStatusCode, "wrong comm status code")
	}

}
