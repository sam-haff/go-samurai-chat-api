package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"go-chat-app-api/internal/accounts"
	"go-chat-app-api/internal/auth"
	"go-chat-app-api/internal/database"
	"go-chat-app-api/internal/middleware"
)

type Server struct {
	Services     Services
	PublicRoutes *gin.RouterGroup
	AuthRoutes   *gin.RouterGroup
	routes       *gin.Engine
}

func (s *Server) Run(addr string) error {
	return s.routes.Run(addr)
}

func Setup(cfg Config) (*Server, error) {
	server := Server{}
	if err := server.Services.Init(cfg); err != nil {
		return nil, err
	}
	server.routes = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.AllowAllOrigins = true
	server.routes.Use(cors.New(corsConfig)) // TODO: change

	// TODO: only inject used deps
	server.routes.Use(middleware.InjectFBApp(server.Services.FirebaseApp), auth.InjectAuth(server.Services.FirebaseAuth), accounts.InjectStorage(server.Services.FirebaseStorage), database.InjectDB(server.Services.MongoDB))

	server.AuthRoutes = server.routes.Group("/", auth.AuthMiddleware)
	server.PublicRoutes = server.routes.Group("/")

	return &server, nil
}

/*func Run(addr string, fbApp *firebase.App, mongoInst *database.MongoDBInstance) error {
	fbAuth := auth.NewAuth(fbApp)
	fbStorage, err := fbApp.Storage(context.TODO())
	if err != nil {
		log.Fatal("Failed to init fb storage. " + err.Error())
	}
	wsHub := websocket.NewWsHub(fbAuth, mongoInst)

	go wsHub.Run()

	routes := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	corsConfig.AllowAllOrigins = true
	routes.Use(cors.New(corsConfig)) // TODO: change

	routes.Use(middleware.InjectFBApp(fbApp), auth.InjectAuth(fbAuth), accounts.InjectStorage(fbStorage), database.InjectDB(mongoInst))
	authRoutes := routes.Group("/", auth.AuthMiddleware)
	publicRoutes := routes.Group("/")
	wsRoutes := routes.Group("/", websocket.InjectWsHub(&wsHub))

	RegisterHandlers(authRoutes, publicRoutes)

	accounts.RegisterHandlers(authRoutes, publicRoutes)
	messages.RegisterHandlers(authRoutes, publicRoutes)
	websocket.RegisterHandlers(authRoutes, publicRoutes, wsRoutes) // TODO: mb pass only ws routes

	return routes.Run(addr)
}*/
