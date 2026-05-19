package api

import (
	"resources_broker/internal/api/handlers"
	"resources_broker/internal/middleware"
	"resources_broker/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Routes(db *mongo.Database, client *asynq.Client) *gin.Engine {
	r := gin.Default()

	taskService := service.NewTaskService(db, client)
	taskHandler := handlers.NewTaskHandler(taskService)

	apiGroup := r.Group("/api")
	{
		auth := apiGroup.Group("/auth")
		{
			auth.POST("/register")
			auth.POST("/login")
			auth.POST("/refresh")
		}

		protected := apiGroup.Group("")
		protected.Use(middleware.JwtMiddleware())
		{
			protected.GET("/auth/me")

			task := protected.Group("/tasks")
			{
				task.POST("", taskHandler.CreateTask)
				task.GET("/:id", taskHandler.GetTask)
				task.GET("", taskHandler.ListTasks)
			}
		}
	}

	return r
}
