package main

import (
	"log"
	"net/http"
	"strconv"

	"nhooyr.io/websocket"
)

var clients map[*websocket.Conn]bool = make(map[*websocket.Conn]bool)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // disable CORS
	})
	if err != nil {
		log.Fatal("Failed to open connection:", err)
	}

	// save the new connected client
	clients[conn] = true

	// notifies all connected clients
	for client := range clients {
		client.Write(r.Context(), websocket.MessageText, []byte("New client connected"))
	}

	for {
		_, data, err := conn.Read(r.Context())
		if err != nil {
			// conn.Write(r.Context(), websocket.MessageText, []byte("Client disconnected"))
			log.Println("Client disconnected")
			delete(clients, conn)
			break
		}

		log.Println(string(data))

		serverResponse := "Received: " + string(data)

		for client := range clients {
			client.Write(r.Context(), websocket.MessageText, []byte(serverResponse))
		}
	}
}

// clients connected
func clientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(strconv.Itoa(len(clients))))
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/clients", clientsHandler)

	http.ListenAndServe(":3000", nil)
}
