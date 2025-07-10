package main

import (
	"log"
	"net/http"

	"go_chat/internal/chat"
	"go_chat/internal/handler"
)

func main() {
	// Cria o hub do chat
	hub := chat.NewHub()
	go hub.Run() // Inicia o hub em sua própria goroutine

	// Configura o manipulador para servir os arquivos estáticos
	fs := http.FileServer(http.Dir("./web/static"))
	http.Handle("/", fs)

	// Configura o manipulador para o endpoint do WebSocket, passando o hub para ele
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWs(hub, w, r)
	})

	log.Println("Servidor iniciado em http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
