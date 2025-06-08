package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"mcphub/internal/handlers"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "mcphub",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/generate-dockerfile", handlers.GenerateDockerfile)
	}

	return router
}
