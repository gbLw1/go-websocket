package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type Client struct {
	Nickname   string
	connection *websocket.Conn
	context    context.Context
}

type Message struct {
	From    string `json:"from"`
	Content string `json:"content"`
	SentAt  string `json:"sentAt"`
}

var (
	clients     map[*Client]bool = make(map[*Client]bool)
	joinCh      chan *Client     = make(chan *Client)
	broadcastCh chan Message     = make(chan Message)
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")

	// open connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // disable CORS
	})
	if err != nil {
		log.Fatal("Server: Failed to open connection:", err)
	}

	go broadcast()
	go joiner()

	// create a client
	client := Client{Nickname: nickname, connection: conn, context: r.Context()}
	joinCh <- &client

	reader(&client)
}

func reader(client *Client) {
	for {
		_, data, err := client.connection.Read(client.context)
		// notifies when client disconnected
		if err != nil {
			log.Println("SERVER: " + client.Nickname + " disconnected")
			delete(clients, client)
			broadcastCh <- Message{
				From:    "SERVER",
				Content: client.Nickname + " disconnected",
				SentAt:  time.Now().Format("02-01-2006 15:04:05"),
			}
			break
		}

		// deserialize message
		var msgReceived Message
		json.Unmarshal(data, &msgReceived)

		// log message to server
		log.Println(msgReceived.From + ": " + msgReceived.Content)

		// broadcast message to all clients
		broadcastCh <- Message{
			From:    msgReceived.From,
			Content: msgReceived.Content,
			SentAt:  time.Now().Format("02-01-2006 15:04:05"),
		}
	}
}

func joiner() {
	// loop while channel is open
	for client := range joinCh {
		clients[client] = true

		log.Println("SERVER: " + client.Nickname + " connected")

		// notifies when a new client connects
		broadcastCh <- Message{
			From:    "SERVER",
			Content: client.Nickname + " connected",
			SentAt:  time.Now().Format("02-01-2006 15:04:05"),
		}
	}
}

func broadcast() {
	for msg := range broadcastCh {
		for client := range clients {
			msg, _ := json.Marshal(msg)
			client.connection.Write(client.context, websocket.MessageText, msg)
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
