package api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Routes(db *mongo.Database) *gin.Engine {
	r := gin.Default()

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	return r
}
