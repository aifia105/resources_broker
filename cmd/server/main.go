package main

import (
	"log"
	"os"
	"resources_broker/internal/api"
	"resources_broker/internal/db"
	"resources_broker/internal/queue"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found, server cannot be started")
	}

	db.ConnectDB()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR not set in .env file")
	}
	client := queue.NewAsynqClient(redisAddr)
	defer client.Close()

	router := api.Routes(db.DB, client)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT not set in .env file")
	}

	log.Printf("Server running on port %s", port)
	log.Printf("---------------------------------------------")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
