package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
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
	setEnvs()

	// Запускаем HTTP-сервер в отдельной горутине
	go startHTTPServer()
	// Запускаем WebSocket-сервер в отдельной горутине
	go startWSServer()

	select {}
}

func startHTTPServer() {
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*.html")
	router.Static("/static", "./web/static")
	router.StaticFile("/favicon.ico", "./web/static/img/favicon.ico")

	router.GET("/mergenator", getMergenatorPage)
	router.POST("/merge", handleMerge)
	router.POST("/webhook/on-push", handleWebhook)

	if err := router.Run("localhost" + HttpPort); err != nil {
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
