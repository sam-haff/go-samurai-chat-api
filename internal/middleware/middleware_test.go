package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	fbauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"go-chat-app-api/internal/auth"
)

func Test_AuthMiddleware(t *testing.T) {
	tests := []struct {
		header             string
		headerVal          string
		token              string
		uid                string
		expectedVerifyCall bool
		expectedInCtx      bool
		status             int
		expectedVerifyErr  error
	}{
		{"Au", "k", "", "someuid1", false, false, http.StatusBadRequest, nil},
		{"Authorization", "Bearer kkk", "kkk", "someuid2", true, false, http.StatusUnauthorized, fmt.Errorf("Error placeholder")},
		{"Authorization", "Bearer ffffff", "ffffff", "someuid3", true, true, http.StatusOK, nil},
	}

	for _, test := range tests {

		var authObj *auth.MockFbAuth = &auth.MockFbAuth{}
		var token fbauth.Token
		token.UID = test.uid
		if test.expectedVerifyCall {
			authObj.On("VerifyToken", mock.Anything, test.token).Return(&token, test.expectedVerifyErr)
		}

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set(test.header, test.headerVal)

		w := httptest.NewRecorder()

		routers := gin.Default()
		routers.Use(InjectParams(nil, authObj, nil))
		routers.GET("/test", AuthMiddleware, func(ctx *gin.Context) {
			v, exists := ctx.Get(CtxVarAuthToken)
			if exists != test.expectedInCtx {
				if exists {
					t.Error("Token expected to be in ctx but is not there")
				} else {
					t.Error("Token expected not to be ctx but it is there")
				}
			}
			_, ok := v.(*fbauth.Token)
			if !ok {
				t.Error("Token is set incorrectly")
			}

			v, exists = ctx.Get(CtxVarUserId)
			if exists != test.expectedInCtx {
				if exists {
					t.Error("Uid expected to be in ctx but is not there")
				} else {
					t.Error("Uid expected not to be ctx but it is there")
				}
			}
			_, ok = v.(string)
			if !ok {
				t.Error("Uid is set incorrectly")
			}
			ctx.String(http.StatusOK, "OK")
		})

		routers.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != test.status {
			t.Errorf("Wrong status code: expected %d, got %d", test.status, resp.StatusCode)
		}

		if test.expectedVerifyCall {
			authObj.AssertCalled(t, "VerifyToken", mock.Anything, test.token)
		}
		fmt.Println(test)
	}
}
