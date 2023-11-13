package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"nhooyr.io/websocket"
)

type Client struct {
	Nickname   string `json:"nickname"`
	connection *websocket.Conn
	context    context.Context
	roomName   string
}

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
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
	roomFromQuery := r.URL.Query().Get("room")

	// validate nickname
	if nickname == "" {
		log.Fatal("Server: No nickname provided")
	}

	// validate room
	if roomFromQuery == "" {
		log.Println("SERVER: No room provided, using default room")
		roomFromQuery = "general"
	}

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
	client := Client{
		Nickname:   nickname,
		connection: conn,
		context:    r.Context(),
		roomName:   roomFromQuery,
	}

	joinCh <- &client

	reader(&client, roomFromQuery)
}

func reader(client *Client, room string) {
	for {
		_, data, err := client.connection.Read(client.context)
		// notifies when client disconnected
		if err != nil {
			log.Println("SERVER: " + client.Nickname + " disconnected from room " + client.roomName)

			delete(clients, client)

			broadcastCh <- Message{
				From:    "SERVER",
				To:      room,
				Content: client.Nickname + " disconnected",
				SentAt:  getTimestamp(),
			}

			break
		}

		// deserialize message
		var msgReceived Message
		json.Unmarshal(data, &msgReceived)

		// log message to server
		log.Println(
			"ROOM: " + msgReceived.To + " -> " + msgReceived.From + ": " + msgReceived.Content,
		)

		// broadcast message to all clients
		broadcastCh <- Message{
			From:    msgReceived.From,
			To:      room,
			Content: msgReceived.Content,
			SentAt:  getTimestamp(),
		}
	}
}

func joiner() {
	// loop while channel is open
	for client := range joinCh {
		clients[client] = true

		log.Println("SERVER: " + client.Nickname + " connected in room " + client.roomName)

		// notifies when a new client connects
		broadcastCh <- Message{
			From:    "SERVER",
			To:      client.roomName,
			Content: client.Nickname + " connected",
			SentAt:  getTimestamp(),
		}
	}
}

func broadcast() {
	for msg := range broadcastCh {
		for client := range clients {
			if client.roomName == msg.To {
				message, _ := json.Marshal(msg)

				client.connection.Write(
					client.context,
					websocket.MessageText,
					message,
				)
			}
		}
	}
}

func clientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var res []*Client
	roomFromQuery := r.URL.Query().Get("room")

	for c := range clients {
		if c.roomName == roomFromQuery {
			res = append(res, c)
		}
	}

	json.NewEncoder(w).Encode(res)
}

func getTimestamp() string {
	return time.Now().UTC().Add(time.Duration(-3) * time.Hour).Format("02-01-2006 15:04:05")
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/clients", clientsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.ListenAndServe("0.0.0.0:"+port, nil)
}
