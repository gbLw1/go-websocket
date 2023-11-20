package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type Client struct {
	ID         string `json:"id"`
	Nickname   string `json:"nickname"`
	Color      string `json:"color"`
	connection *websocket.Conn
	context    context.Context
	roomName   string
}

const (
	// Message types
	MESSAGE      = "message"
	NOTIFICATION = "notification"
)

type Message struct {
	Type     string `json:"type"` // message or notification
	From     Client `json:"from"`
	to       string
	Content  string `json:"content"`  // required in Type: message
	IsTyping bool   `json:"isTyping"` // required in Type: notification
	SentAt   string `json:"sentAt"`
}

var (
	clients     map[*Client]bool = make(map[*Client]bool)
	joinCh      chan *Client     = make(chan *Client)
	broadcastCh chan Message     = make(chan Message)
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")
	room := r.URL.Query().Get("room")

	// validate nickname
	if nickname == "" {
		log.Fatal("Server: No nickname provided")
	}

	// validate room
	if room == "" {
		log.Println("SERVER: No room provided, using default room")
		room = "general"
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
		ID:         uuid.New().String(),
		Nickname:   nickname,
		connection: conn,
		context:    r.Context(),
		roomName:   room,
	}

	joinCh <- &client

	reader(&client, room)
}

func reader(client *Client, room string) {
	for {
		_, data, err := client.connection.Read(client.context)
		// notifies when client disconnected
		if err != nil {
			log.Println("SERVER: " + client.Nickname + " disconnected from room " + client.roomName)

			broadcastCh <- Message{
				Type:     NOTIFICATION,
				From:     Client{Nickname: client.Nickname},
				to:       room,
				IsTyping: false,
				SentAt:   getTimestamp(),
			}

			delete(clients, client)
			client.connection.Close(websocket.StatusNormalClosure, "")

			broadcastCh <- Message{
				Type:    MESSAGE,
				From:    Client{Nickname: "SERVER", Color: "#64BFFF"},
				to:      room,
				Content: client.Nickname + " disconnected",
				SentAt:  getTimestamp(),
			}

			break
		}

		// deserialize message
		var msgReceived Message
		json.Unmarshal(data, &msgReceived)

		// log message to server
		if msgReceived.Type == MESSAGE {
			log.Printf("ROOM: %s -> %s: %s\n", room, msgReceived.From.Nickname, msgReceived.Content)
		} else if msgReceived.Type == NOTIFICATION {
			log.Printf("ROOM: %s -> %s is typing: %t\n", room, msgReceived.From.Nickname, msgReceived.IsTyping)
		} else {
			log.Printf("ROOM: %s -> %s sent an unknown message type\n", room, msgReceived.From.Nickname)
			continue
		}

		// broadcast message to all clients
		broadcastCh <- Message{
			Type: msgReceived.Type,
			From: Client{
				ID:       client.ID,
				Nickname: client.Nickname,
				Color:    msgReceived.From.Color,
			},
			to:       room,
			Content:  msgReceived.Content,
			IsTyping: msgReceived.IsTyping,
			SentAt:   getTimestamp(),
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
			Type:    MESSAGE,
			From:    Client{Nickname: "SERVER", Color: "#64BFFF"},
			to:      client.roomName,
			Content: client.Nickname + " connected",
			SentAt:  getTimestamp(),
		}
	}
}

func broadcast() {
	for msg := range broadcastCh {
		for client := range clients {
			if client.roomName == msg.to {
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
