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

type Message struct {
	From    Client `json:"from"`
	to      string
	Content string `json:"content"`
	SentAt  string `json:"sentAt"`
}

type Typing struct {
	Client   string `json:"client"`
	room     string
	IsTyping bool `json:"isTyping"`
}

var (
	clients     map[*Client]bool = make(map[*Client]bool)
	joinCh      chan *Client     = make(chan *Client)
	broadcastCh chan Message     = make(chan Message)
	typingCh    chan Typing      = make(chan Typing)
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

			delete(clients, client)
			client.connection.Close(websocket.StatusNormalClosure, "")

			broadcastCh <- Message{
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

		if client.Color == "" && msgReceived.From.Color != "" {
			client.Color = msgReceived.From.Color
		}

		// log message to server
		log.Println(
			"ROOM: " + room + " -> " + msgReceived.From.Nickname + ": " + msgReceived.Content,
		)

		// broadcast message to all clients
		broadcastCh <- Message{
			From:    msgReceived.From,
			to:      room,
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

func notifyTypingHandler(w http.ResponseWriter, r *http.Request) {
	room := r.URL.Query().Get("room")
	if room == "" {
		log.Fatal("SERVER: No room provided to listen to typing")
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Fatal("SERVER: Failed to open connection:", err)
	}

	go typingBroadcast(conn, r.Context())

	client := Client{
		ID:         uuid.New().String(),
		connection: conn,
		context:    r.Context(),
		roomName:   room,
	}

	readerTyping(&client, room)
}

func readerTyping(client *Client, room string) {
	for {
		_, data, err := client.connection.Read(client.context)
		if err != nil {
			log.Println("SERVER: " + client.Nickname + " disconnected from room " + client.roomName)

			delete(clients, client)
			client.connection.Close(websocket.StatusNormalClosure, "")

			break
		}

		var typingInfo Typing
		json.Unmarshal(data, &typingInfo)

		if typingInfo.IsTyping {
			log.Println(
				"ROOM: " + room + " -> " + typingInfo.Client + " is typing...",
			)
		} else {
			log.Println("Room: " + room + " -> " + typingInfo.Client + " stopped typing")
		}

		typingCh <- Typing{
			Client:   typingInfo.Client,
			room:     room,
			IsTyping: typingInfo.IsTyping,
		}
	}
}

func typingBroadcast(conn *websocket.Conn, ctx context.Context) {
	for typer := range typingCh {
		for client := range clients {
			if client.roomName == typer.room {
				json, _ := json.Marshal(typer)

				conn.Write(
					ctx,
					websocket.MessageText,
					json,
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
	http.HandleFunc("/typing", notifyTypingHandler)
	http.HandleFunc("/clients", clientsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.ListenAndServe("0.0.0.0:"+port, nil)
}
