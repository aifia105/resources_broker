package main

import (
	"log"
	"os"
	"resources_broker/internal/api"
	"resources_broker/internal/db"

	"github.com/joho/godotenv"
)

func main() {
	db.ConnectDB()

	router := api.Routes(db.DB)

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found, server cannot be started")
	}

	if err := router.Run(":" + os.Getenv("PORT")); err != nil {
		log.Fatal("Failed to start server:", err)
	}
	log.Printf("Server running on port %s", os.Getenv("PORT"))
}
