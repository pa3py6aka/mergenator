package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

var (
	client = &http.Client{Timeout: 10 * time.Second}
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*.html")
	router.Static("/static", "./web/static")

	router.GET("/mergenator", getMergenatorPage)
	router.POST("/merge", handleMerge)

	if err := router.Run("localhost:8080"); err != nil {
		panic(err)
	}
}

func getMergenatorPage(c *gin.Context) {
	c.HTML(200, "main.html", gin.H{})
}
