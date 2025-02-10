package websocket

import (
	"io"
	"log"
	"net/http"
	"os"
)

func queryPresenceOnlineStatus(uid string) bool {
	presencePwd := os.Getenv("PRESENCE_PWD")

	// TODO: load url from env
	req, err := http.NewRequest("GET", "http://presence:8080/online/"+uid, nil)
	req.Header.Set("Authorization", presencePwd)

	if err != nil {
		log.Printf("Failed to create presence req with: %s", err.Error())
		return false
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to query presence req with: %s", err.Error())
		return false
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read presence responce: %s", err.Error())
		return false
	}

	if len(b) == 0 {
		log.Printf("Presence responce is ill formed")
		return false
	}

	return int(b[0]) > 0
}
