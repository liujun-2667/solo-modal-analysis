package main

import (
	"modal-analysis/internal/handler"
	"modal-analysis/internal/preset"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})

	preset.LoadPresets()

	r.POST("/api/analyze", handler.AnalyzeModal)
	r.GET("/api/presets", handler.GetPresets)
	r.POST("/api/preset/:name", handler.LoadPreset)
	r.POST("/api/frf", handler.CalculateFRF)
	r.POST("/api/transient", handler.CalculateTransient)

	r.Run(":8080")
}
