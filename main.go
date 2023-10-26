package main

import (
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // disable CORS
	})
	if err != nil {
		log.Fatal("Failed to open connection:", err)
	}

	for {
		_, data, err := conn.Read(r.Context())
		if err != nil {
			log.Println(err)
			break
		}

		log.Println(string(data))

		serverResponse := "Received: " + string(data)

		conn.Write(r.Context(), websocket.MessageText, []byte(serverResponse))
	}
}

func main() {
	http.HandleFunc("/", wsHandler)

	http.ListenAndServe(":3000", nil)
}
