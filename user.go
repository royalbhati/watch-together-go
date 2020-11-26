package main

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// client represents a single chatting user.
type user struct {
	name string

	id uuid.UUID

	color string

	// socket is the web socket for this client.
	socket *websocket.Conn

	// send is a channel on which messages are sent.
	send chan []byte

	// room is the room this client is chatting in.
}

//Reads the message from the socket
func (u *user) read(h *hall) {
	defer u.socket.Close()
	for {
		_, msg, err := u.socket.ReadMessage()
		// fmt.Printf("msg reciebed")
		if err != nil {
			return
		}
		//passing that message to the forward channel
		h.forward <- msg
	}
}

func (u *user) write() {
	defer u.socket.Close()
	for msg := range u.send {
		// fmt.Printf("%s", "WRITE FUNC")

		// fmt.Printf("%s", msg)
		// fmt.Printf("%s", "WRITE FUNC END")
		// fmt.Printf("%+v", u)

		err := u.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
