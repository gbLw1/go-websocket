package main

import (
	"encoding/json"
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

type Client struct {
	Nickname   string
	connection *websocket.Conn
}

var clients map[*Client]bool = make(map[*Client]bool)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // disable CORS
	})
	if err != nil {
		log.Fatal("Server: Failed to open connection:", err)
	}

	client := Client{
		Nickname:   nickname,
		connection: conn,
	}
	clients[&client] = true

	// notifies all connected clients
	for c := range clients {
		c.connection.Write(
			r.Context(),
			websocket.MessageText,
			[]byte("Server: User '"+nickname+"' connected"),
		)
	}

	for {
		_, data, err := client.connection.Read(r.Context())
		if err != nil {
			// conn.Write(r.Context(), websocket.MessageText, []byte("Client disconnected"))
			log.Println("Server: " + nickname + "disconnected")
			delete(clients, &client)
			break
		}

		log.Println(string(data))

		serverResponse := "Received: " + string(data)

		for client := range clients {
			client.connection.Write(r.Context(), websocket.MessageText, []byte(serverResponse))
		}
	}
}

func clientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var res []*Client
	for c := range clients {
		res = append(res, c)
	}
	json.NewEncoder(w).Encode(res)
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/clients", clientsHandler)

	http.ListenAndServe(":42069", nil)
}
