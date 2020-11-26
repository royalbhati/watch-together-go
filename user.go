package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type user struct {
	name string

	id uuid.UUID

	color string

	socket *websocket.Conn

	send chan []byte
}

//Reads the message from the socket
func (u *user) read(h *hall) {
	defer u.socket.Close()
	for {
		_, msg, err := u.socket.ReadMessage()
		if err != nil {
			return
		}
		h.forward <- msg
	}
}

func (u *user) write() {
	defer u.socket.Close()
	for msg := range u.send {

		err := u.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
