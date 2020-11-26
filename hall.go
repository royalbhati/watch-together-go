package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/matryer/goblueprints/chapter1/trace"
)

const (
	NewRoom    = "newRoom"
	AssignID   = "assignID"
	JoinRoom   = "joinRoom"
	NewMessage = "newMessage"
)

type hall struct {

	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.

	socket *websocket.Conn

	forward chan []byte

	// join is a channel for clients wishing to join the room.
	join chan *user

	// leave is a channel for users wishing to leave the room.
	leave chan *user

	// clients holds all current clients in this room.
	rooms map[uuid.UUID]*room
	users map[uuid.UUID]*user

	tracer trace.Tracer
}

// newRoom makes a new room that is ready to
// go.
func newHall() *hall {
	return &hall{
		forward: make(chan []byte),
		join:    make(chan *user),
		leave:   make(chan *user),
		rooms:   make(map[uuid.UUID]*room),
		users:   make(map[uuid.UUID]*user),
		tracer:  trace.Off(),
	}
}

func marshalData(userData map[string]interface{}) []byte {
	val, err := json.Marshal(userData)

	//What to do in the case of an error?
	if err != nil {
		return val
	}
	return nil
}
func sendData(client *user, val []byte) {
	client.send <- val
}

func (h *hall) runHall() {
	for {
		select {
		case client := <-h.join:
			// joining
			//handle errors make one function
			h.tracer.Trace("New client joined")
			h.users[client.id] = client
			fmt.Printf("%+v\n", client)
			userData := map[string]interface{}{
				"type": AssignID,
				"id":   client.id.String(),
			}
			val := marshalData(userData)
			sendData(client, val)

		case client := <-h.leave:

			// leaving
			// delete(h.clients, client)
			payload := map[string]interface{}{
				"type": "disconnect",
				"id":   client.id,
			}
			val := marshalData(payload)
			h.forward <- val
			close(client.send)
			h.tracer.Trace("Client left")
		case msg := <-h.forward:
			h.tracer.Trace("Message received: ", string(msg))
			var data socketData
			json.Unmarshal([]byte(msg), &data)
			fmt.Println("heyyyy", data.Type)
			switch data.Type {
			case "create":
				roomID := uuid.New()
				host := h.users[data.HostID]
				usersinRoom := make(map[uuid.UUID]*user)
				(*host).name = data.HostName
				usersinRoom[data.HostID] = host
				h.rooms[roomID] = &room{
					id:       roomID,
					hostName: data.HostName,
					vidID:    data.VidID,
					hostID:   data.HostID,
					users:    usersinRoom,
					roomSize: 2,
				}

				newData := make(map[string]interface{})
				a := make([]string, 0)
				for _, v := range usersinRoom {
					a = append(a, v.name)

				}
				newData["type"] = NewRoom
				newData["users"] = a
				newData["roomID"] = roomID
				newData["vId"] = data.VidID
				newData["host"] = data.HostName
				newData["name"] = data.HostName

				fmt.Println("------")
				fmt.Printf("%+v\n", a)
				fmt.Println("------")
				val, _ := json.Marshal(newData)
				host.send <- val
			case "join":
				joinee := h.users[data.JoineeID]
				room := h.rooms[data.RoomID]
				room.users[data.JoineeID] = joinee
				(*joinee).name = data.UserName

				newData := make(map[string]interface{})
				a := make([]string, 0)
				for _, v := range room.users {
					if v.name != "" {
						a = append(a, v.name)
					}

				}

				//SEND A PING TOEACH USER TO UPDATE THEIR LIST
				// joinee.send <- val

				newData["type"] = JoinRoom
				newData["users"] = a
				newData["roomID"] = data.RoomID
				newData["vId"] = room.vidID
				newData["host"] = room.hostName
				newData["name"] = data.UserName

				fmt.Println("---JOIN---")
				fmt.Printf("%+v\n", a)
				fmt.Println("------")
				val, _ := json.Marshal(newData)
				joinee.send <- val
			case "vid":
				room := h.rooms[data.RoomID]
				newData := make(map[string]interface{})
				newData["type"] = "vidChange"
				newData["status"] = data.Status
				newData["roomID"] = data.RoomID

				//hack to see if empty value
				newData["time"] = data.Time

				val, _ := json.Marshal(newData)

				//handle closed channels
				// by removing guys once connection closed
				for _, v := range room.users {
					// if v.id != data.JoineeID {
					v.send <- val
					// }
				}

			case "message":
				room := h.rooms[data.RoomID]
				sender := room.users[data.JoineeID]
				newData := make(map[string]interface{})
				newData["type"] = "newMessage"
				newData["text"] = data.Text
				newData["sender"] = sender.name

				val, _ := json.Marshal(newData)

				//handle closed channels
				// by removing guys once connection closed
				for _, v := range room.users {
					// if v.id != data.JoineeID {
					v.send <- val
					// }
				}
			case "disconnect":
				room := h.rooms[data.RoomID]

				user, ok := room.users[data.JoineeID]
				if ok {
					delete(room.users, data.JoineeID)
				}

				a := make([]string, 0)
				for _, v := range room.users {
					a = append(a, v.name)

				}
				newData := make(map[string]interface{})
				newData["type"] = "userLeft"
				newData["id"] = user.name
				newData["users"] = a
				val, _ := json.Marshal(newData)
				for _, v := range room.users {
					v.send <- val
				}
			default:
				fmt.Println("Can I get a ohhh yeahhhh!")
			}
			h.tracer.Trace(" -- sent to client")

		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (h *hall) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	h.socket = socket
	newUser := &user{
		socket: socket,
		id:     uuid.New(),
		color:  "red",
		send:   make(chan []byte, messageBufferSize),
	}

	h.join <- newUser
	defer func() { h.leave <- newUser }()
	go newUser.write()
	newUser.read(h)
}
