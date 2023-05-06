package main

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type empty struct{}

var null = empty{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

var clients = map[*Client]empty{}
var hosts = map[*Client]empty{}

var done = make(chan empty)
var allowedTypes = map[string]empty{
	"client": null,
	"host":   null,
}

var hub *Hub

func upgrade(w http.ResponseWriter, r *http.Request) {
	var t []string
	var ok bool
	if t, ok = r.URL.Query()["type"]; !ok {
		return
	}

	connType := ConnectionTypeFromString(t[0])
	if connType == UnknownConn {
		log.Printf("Unknown connection type: %v", r.URL.String())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	client := &Client{
		connectionType: connType,
		conn:           conn,
		hub:            hub,
		send:           make(chan []byte),
	}
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}

func broadcast(receivers map[*websocket.Conn]empty, messType int, raw []byte) {
	for conn := range receivers {
		conn.WriteMessage(messType, raw)
	}
}

func main() {
	hub = NewHub()
	go hub.Run()
	defer func() {
		close(hub.close)
	}()

	http.HandleFunc("/", upgrade)

	log.Println("Listening on 0.0.0.0:4001")
	listener, err := net.Listen("tcp4", "0.0.0.0:4001")
	if err != nil {
		panic(err)
	}
	http.Serve(listener, nil)
}
