package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

const (
	WSMessageTypeDefault = "default"
	WSMessageTypeError   = "error"
	WSMessageTypeSuccess = "success"
	WSMessageTypeHeader  = "header"
)

var (
	allowedOrigins = map[string]bool{
		"http://localhost:8080": true,
		WSAllowedOrigin:         true,
	}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return allowedOrigins[origin]
		},
	}
)

type (
	WSMessage struct {
		Message string `json:"message,omitempty"`
		Type    string `json:"type,omitempty"`
	}
	Client struct {
		conn *websocket.Conn
		id   string
	}
)

var clients = make(map[string]*Client) // ID → Client
var mutex = sync.Mutex{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка установки WebSocket:", err)
		return
	}

	// Генерируем уникальный ID
	clientID := uuid.New().String()

	client := &Client{
		conn: conn,
		id:   clientID,
	}

	// Сохраняем клиента
	mutex.Lock()
	clients[clientID] = client
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(clients, clientID)
		mutex.Unlock()
		conn.Close()
	}()

	// Отправляем ID клиенту сразу после подключения
	err = conn.WriteMessage(websocket.TextMessage, []byte("{\"clientID\":\""+clientID+"\"}"))
	if err != nil {
		log.Printf("Ошибка отправки ID клиенту: %v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Соединение закрыто:", err)
			break
		}
		log.Printf("Получено от %s: %s", clientID, message)
	}
}

func sendMessageByID(id string, message string, WSMessageType string) {
	mutex.Lock()
	client, exists := clients[id]
	mutex.Unlock()

	if !exists {
		log.Printf("WS-Клиент %s не найден", id)
		return
	}

	wsMessage := WSMessage{
		Message: message,
		Type:    WSMessageType,
	}

	jsonData, err := json.Marshal(wsMessage)
	if err != nil {
		log.Println("Ошибка сериализации:", err)
		return
	}

	err = client.conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		log.Print("ошибка отправки соообщения по ws")
	}
}

// Функция для отправки сообщения всем подключённым клиентам
func broadcastMessage(message []byte) {
	mutex.Lock()
	// Копируем соединения в отдельный срез (чтобы не держать mutex при отправке)
	currentClients := make([]*websocket.Conn, 0, len(clients))
	for _, client := range clients {
		currentClients = append(currentClients, client.conn)
	}
	mutex.Unlock()

	// Отправляем сообщение каждому клиенту
	for _, conn := range currentClients { // Используем _, conn (а не conn := range)
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Ошибка отправки сообщения клиенту: %v", err)

			// Ищем и удаляем клиента по соединению
			mutex.Lock()
			for id, client := range clients {
				if client.conn == conn {
					client.conn.Close()
					delete(clients, id)
					log.Printf("Клиент с ID %s удалён из‑за ошибки отправки", id)
					break
				}
			}
			mutex.Unlock()
		}
	}
}

func startWSServer() {
	http.HandleFunc("/ws", wsHandler)
	log.Printf("WebSocket-сервер запущен на %s", WSPort)
	if err := http.ListenAndServe(WSPort, nil); err != nil {
		log.Fatalf("Ошибка WebSocket-сервера: %v", err)
	}
}
