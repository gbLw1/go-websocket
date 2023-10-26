package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type Client struct {
	Nickname   string
	connection *websocket.Conn
}

type Message struct {
	From    string `json:"from"`
	Content string `json:"content"`
	SentAt  string `json:"sentAt"`
}

var clients map[*Client]bool = make(map[*Client]bool)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")

	// open connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // disable CORS
	})
	if err != nil {
		log.Fatal("Server: Failed to open connection:", err)
	}

	// create a client
	client := Client{
		Nickname:   nickname,
		connection: conn,
	}
	clients[&client] = true

	// notifies chat that a new client connected
	for c := range clients {
		msg, _ := json.Marshal(Message{
			From:    "SERVER",
			Content: nickname + " connected",
			SentAt:  time.Now().Format("02-01-2006 15:04:05"),
		})

		c.connection.Write(r.Context(), websocket.MessageText, msg)
	}

	for {
		// read client messages
		_, data, err := client.connection.Read(r.Context())
		if err != nil {
			log.Println("SERVER: " + nickname + " disconnected")
			delete(clients, &client)

			// notifies chat that client disconnected
			for c := range clients {
				msg, _ := json.Marshal(Message{
					From:    "SERVER",
					Content: nickname + " disconnected",
					SentAt:  time.Now().Format("02-01-2006 15:04:05"),
				})

				c.connection.Write(r.Context(), websocket.MessageText, msg)
			}

			break
		}

		// deserialize message
		var msgReceived Message
		json.Unmarshal(data, &msgReceived)

		// log message to server
		log.Println(msgReceived.From + ": " + msgReceived.Content)

		// write client messages
		for client := range clients {
			msg, err := json.Marshal(Message{
				From:    msgReceived.From,
				Content: msgReceived.Content,
				SentAt:  time.Now().Format("02-01-2006 15:04:05"),
			})
			if err != nil {
				log.Println("SERVER: Failed to serialize message:", err)
				continue
			}

			client.connection.Write(r.Context(), websocket.MessageText, msg)
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

	http.ListenAndServe(":1337", nil)
}
