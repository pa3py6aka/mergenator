package main

import (
	"crypto/tls"
	"github.com/gin-gonic/gin"
	"log"
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
	// go startWSServer()

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

	router.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})

	// Создаём TLS-конфигурацию
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Минимальная версия TLS
	}

	if OverProxy {
		if err := router.Run("localhost" + HttpPort); err != nil {
			panic(err)
		}
	} else {
		// Создаём HTTP-сервер с TLS-конфигурацией
		server := &http.Server{
			Addr:      "localhost" + HttpPort, // например, ":8085"
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		log.Printf("HTTPS server running on https://localhost%s", HttpPort)
		err := server.ListenAndServeTLS(SSLCertPem, SSLKeyPem)
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTPS server failed to start: %v", err)
		}
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
