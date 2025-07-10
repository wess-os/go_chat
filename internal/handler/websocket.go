package handler

import (
	"log"
	"net/http"
	"time"

	"go_chat/internal/chat"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// readPump bombeia mensagens do websocket para o hub.
func readPump(client *chat.Client) {
	defer func() {
		client.Hub.Unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error { client.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		client.Hub.Broadcast <- message
	}
}

// writePump bombeia mensagens do hub para o websocket.
func writePump(client *chat.Client) {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.Conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWs(hub *chat.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &chat.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	go writePump(client)
	go readPump(client)
}
