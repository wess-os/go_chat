package chat

import "github.com/gorilla/websocket"

type Client struct {
	Hub *Hub

	// A conexão websocket.
	Conn *websocket.Conn

	// Canal de buffer de mensagens de saída.
	Send chan []byte
}
