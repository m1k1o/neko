package main

import (
	"log"
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

var clients = map[*websocket.Conn]empty{}
var hosts = map[*websocket.Conn]empty{}

var done = make(chan empty)
var allowedTypes = map[string]empty{
	"client": null,
	"host":   null,
}

func contains(haystack map[string]empty, needle string) bool {
	_, ok := haystack[needle]

	return ok
}

func upgrade(w http.ResponseWriter, r *http.Request) {
	var t []string
	var ok bool
	if t, ok = r.URL.Query()["type"]; !ok || !contains(allowedTypes, t[0]) {
		return
	}

	connType := t[0]
	log.Println(connType)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	targets := hosts
	receivers := clients
	if connType == "client" {
		targets = clients
		receivers = hosts
	}

	targets[conn] = null
	log.Printf("New connection %s: %d", connType, len(targets))

	defer func() {
		defer conn.Close()
		delete(targets, conn)
		log.Printf("Number of %s Connections: %d", connType, len(targets))
	}()

	for {
		messType, raw, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}

		broadcast(receivers, messType, raw)
	}
}

func broadcast(receivers map[*websocket.Conn]empty, messType int, raw []byte) {
	for conn := range receivers {
		conn.WriteMessage(messType, raw)
	}
}

func main() {
	defer func() {
		for conn := range hosts {
			conn.Close()
		}
		for conn := range clients {
			conn.Close()
		}
	}()

	http.HandleFunc("/", upgrade)
	log.Println("Listening on 0.0.0.0:4001")
	log.Fatal(http.ListenAndServe("0.0.0.0:4001", nil))
}
