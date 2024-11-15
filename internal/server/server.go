package server

import (
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/messages"
	"go-chat-app-api/internal/middleware"
	"go-chat-app-api/internal/users"
)

func Run(addr string, fbApp *firebase.App, mongoInst *database.MongoDBInstance) error {
	routers := gin.Default()

	routers.Use(middleware.InjectParams(fbApp, mongoInst))

	auth.RegisterHandlers(routers)
	users.RegisterHandlers(routers)
	messages.RegisterHandlers(routers)

	routers.Run(addr)

	return nil
}
