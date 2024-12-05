package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

// upgrader for Websocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// store connected clients
var clients = make(map[*websocket.Conn]bool)

// broadcast channel
var broadcast = make(chan Message)

// struct for message
type Message struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

func main() {
	http.HandleFunc("/ws", handleConnections)

	// start listening for incoming chat messages
	go handleMessages()

	fmt.Println("server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("failed to start server: %v", err)
	}
}

// WebSocket connection handler
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("failed to upgrade Websocket", err)
		return
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("failed to load messages:", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

// Message handler
func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println("fail to broadcast:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
