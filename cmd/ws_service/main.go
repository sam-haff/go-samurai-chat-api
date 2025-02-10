// handling realtime chat communication
package main

import (
	"go-chat-app-api/internal/server"
	"go-chat-app-api/internal/websocket"
	"log"
)

func Run() error {
	cfg := server.ReadConfigFromEnv()
	cfg.RequireNATS = true

	server, err := server.Setup(cfg)
	if err != nil {
		return err
	}

	wsHub := websocket.NewWsHub(server.Services.FirebaseAuth, server.Services.MongoDB, server.Services.Nats)
	go wsHub.Run()

	server.PublicRoutes.Use(websocket.InjectWsHub(&wsHub))
	websocket.RegisterHandlers(nil, nil, server.PublicRoutes)

	return server.Run(":8080")
}

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Failed to start WebSocket/Chat Hub server: %s", err.Error())
	}
}
