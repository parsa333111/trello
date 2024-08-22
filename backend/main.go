package main

import (
	"log"
	"os"

	"github.com/skye-tan/trello/backend/database"
	endpoints "github.com/skye-tan/trello/backend/endpoints"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
	"github.com/skye-tan/trello/backend/websocket_utils"
)

func main() {
	log_file, err := os.OpenFile("backend.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer log_file.Close()

	log.SetOutput(log_file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listen_address, ok := os.LookupEnv("LISTEN_ADDRESS")
	if !ok {
		log.Println("Warn: Missing enviroment variable LISTEN_ADDRESS.",
			"Using default listen address: [0.0.0.0:8081]")
		listen_address = "0.0.0.0:8081"
	}

	database.InitializeDatabase()

	monitoring.InitalizeStatistics()

	go websocket_utils.Hub.Run()

	endpoints.Start(listen_address)
}
