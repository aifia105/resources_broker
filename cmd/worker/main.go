package main

import (
	"log"
	"os"
	"os/signal"
	"resources_broker/internal/db"
	"resources_broker/internal/queue"
	"resources_broker/internal/service"
	workerhandler "resources_broker/internal/worker"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"

	"strconv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found, worker cannot be started")
	}

	db.ConnectDB()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR not set in .env file")
	}

	concurrency := 10
	if concurrencyStr := os.Getenv("WORKER_CONCURRENCY"); concurrencyStr != "" {
		if c, err := strconv.Atoi(concurrencyStr); err == nil && c > 0 {
			concurrency = c
		}
	}

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: concurrency},
	)

	mux := asynq.NewServeMux()
	provisionService := service.NewProvisionResourceService(db.DB)
	taskService := service.NewTaskService(db.DB, nil)
	provisionHandler := workerhandler.NewProvisionHandler(provisionService, taskService)
	mux.HandleFunc(queue.TypeProvisionResource, provisionHandler.HandleProvisionTask)

	go func() {
		if err := server.Run(mux); err != nil {
			log.Fatalf("Failed to run worker server: %v", err)
		}
	}()

	log.Println("Worker running and waiting for queue jobs")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	server.Shutdown()
	log.Println("Worker stopped")
}
