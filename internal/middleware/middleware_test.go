package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_InjectFBApp(t *testing.T) {
	assert := assert.New(t)

	testFbApp := &firebase.App{}
	tests := []struct {
		val    *firebase.App
		exists bool
	}{
		{nil, false},
		{testFbApp, true},
	}

	for _, test := range tests {
		routes := gin.Default()
		if test.exists {
			routes.Use(InjectFBApp(testFbApp))
		}
		routes.GET("/test", func(ctx *gin.Context) {
			v, exists := ctx.Get(CtxVarFirebaseApp)

			assert.Equal(test.exists, exists, "wrong existance")

			if exists {
				fbApp := v.(*firebase.App)
				assert.Equal(test.val, fbApp, "wrong value in the context")
			}
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		routes.ServeHTTP(rec, req)
	}
}
