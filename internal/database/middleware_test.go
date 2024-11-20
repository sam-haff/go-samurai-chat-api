package database

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_InjectDB(t *testing.T) {
	assert := assert.New(t)

	testDb := &MongoDBInstance{}
	tests := []struct {
		val    *MongoDBInstance
		exists bool
	}{
		{nil, false},
		{testDb, true},
	}

	for _, test := range tests {
		routes := gin.Default()
		if test.exists {
			routes.Use(InjectDB(test.val))
		}
		routes.GET("/test", func(ctx *gin.Context) {
			v, exists := ctx.Get(CtxVarMongoDBInst)

			assert.Equal(test.exists, exists, "shouldnt or should exist")

			if exists {
				fbApp := v.(*MongoDBInstance)
				assert.Equal(test.val, fbApp, "wrong value in the context")
			}
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		routes.ServeHTTP(rec, req)
	}
}
