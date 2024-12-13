package server

import (
	firebase "firebase.google.com/go/v4"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/messages"
	"go-chat-app-api/internal/middleware"
	"go-chat-app-api/internal/websocket"
)

func Run(addr string, fbApp *firebase.App, mongoInst *database.MongoDBInstance) error {
	fbAuth := auth.NewAuth(fbApp)
	wsHub := websocket.NewWsHub(fbAuth, mongoInst)

	go wsHub.Run()

	routes := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.AllowAllOrigins = true
	routes.Use(cors.New(corsConfig)) // TODO: change

	routes.Use(middleware.InjectFBApp(fbApp), auth.InjectAuth(fbAuth), database.InjectDB(mongoInst))
	authRoutes := routes.Group("/", auth.AuthMiddleware)
	publicRoutes := routes.Group("/")
	wsRoutes := routes.Group("/", websocket.InjectWsHub(&wsHub))

	RegisterHandlers(authRoutes, publicRoutes)

	accounts.RegisterHandlers(authRoutes, publicRoutes)
	messages.RegisterHandlers(authRoutes, publicRoutes)
	websocket.RegisterHandlers(authRoutes, publicRoutes, wsRoutes) // TODO: mb pass only ws routes

	return routes.Run(addr)
}
