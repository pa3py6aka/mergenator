package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	HTTP_PORT = ":8080" // Порт для HTTP-сервера
	WS_PORT   = ":8090" // Порт для WebSocket-сервера
)

var (
	httpClient = &http.Client{Timeout: 10 * time.Second}
)

type Repository struct {
	StandBranch string
	ProjectId   string
	AssigneeId  int
}

func main() {
	// Запускаем HTTP-сервер в отдельной горутине
	go startHTTPServer()

	// Запускаем WebSocket-сервер в отдельной горутине
	go startWSServer()

	// Ждём завершения (на практике — сервер работает бесконечно)
	select {}
}

func startHTTPServer() {
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
	_, err := c.Cookie("gitlab_user_id")
	if err != nil {
		c.HTML(200, "login.html", gin.H{})
		return
	}

	c.HTML(200, "main.html", gin.H{})
}
