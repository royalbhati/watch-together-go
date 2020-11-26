package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	NewRoom    = "newRoom"
	AssignID   = "assignID"
	JoinRoom   = "joinRoom"
	NewMessage = "newMessage"
)

type hall struct {
	socket *websocket.Conn

	forward chan []byte

	join chan *user

	leave chan *user

	rooms map[uuid.UUID]*room

	users map[uuid.UUID]*user
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
	}
}

func marshalData(userData map[string]interface{}) []byte {
	val, err := json.Marshal(userData)

	//What to do in the case of an error?
	//Donnt want to exit because webserver
	if err != nil {
		return nil
	}
	return val
}
func sendData(client *user, val []byte) {
	client.send <- val
}

func sendToAll(users map[uuid.UUID]*user, val []byte) {
	for _, v := range users {
		v.send <- val
	}

}

func (h *hall) runHall() {
	for {
		select {
		case client := <-h.join:
			// joining
			//handle errors make one function
			h.users[client.id] = client
			fmt.Printf("%+v\n", client)
			userData := map[string]interface{}{
				"type": AssignID,
				"id":   client.id.String(),
			}
			val := marshalData(userData)
			sendData(client, val)

		case client := <-h.leave:

			payload := map[string]interface{}{
				"type": "disconnect",
				"id":   client.id,
			}
			val := marshalData(payload)
			h.forward <- val
			close(client.send)
		case msg := <-h.forward:
			var data socketData
			json.Unmarshal([]byte(msg), &data)
			fmt.Println("event", string(msg))
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
				a := make([]string, 0)
				for _, v := range usersinRoom {
					a = append(a, v.name)

				}
				payload := map[string]interface{}{
					"type":   NewRoom,
					"users":  a,
					"roomID": roomID,
					"vId":    data.VidID,
					"host":   data.HostName,
					"name":   data.HostName,
				}

				val := marshalData(payload)
				sendData(host, val)
			case "join":
				joinee := h.users[data.JoineeID]
				room := h.rooms[data.RoomID]
				room.users[data.JoineeID] = joinee
				(*joinee).name = data.UserName

				a := make([]string, 0)
				for _, v := range room.users {
					if v.name != "" {
						a = append(a, v.name)
					}

				}
				payload := map[string]interface{}{
					"type":   JoinRoom,
					"users":  a,
					"roomID": data.RoomID,
					"vId":    room.vidID,
					"host":   room.hostName,
					"name":   data.UserName,
				}
				notifyPayload := map[string]interface{}{
					"type":  "updateUsers",
					"users": a,
				}

				val := marshalData(payload)
				notifyOthers := marshalData(notifyPayload)
				sendData(joinee, val)
				sendToAll(room.users, notifyOthers)

			case "vid":
				room := h.rooms[data.RoomID]
				payload := map[string]interface{}{
					"type":   "vidChange",
					"status": data.Status,
					"roomID": data.RoomID,
					"time":   data.Time,
				}
				val := marshalData(payload)
				sendToAll(room.users, val)

			case "message":
				room := h.rooms[data.RoomID]
				sender := room.users[data.JoineeID]
				payload := map[string]interface{}{
					"type":   "newMessage",
					"text":   data.Text,
					"sender": sender.name,
				}
				val := marshalData(payload)
				sendToAll(room.users, val)

			case "changeVideo":
				room := h.rooms[data.RoomID]
				room.vidID = data.VidID
				payload := map[string]interface{}{
					"type": "newVideo",
					"vId":  data.VidID,
				}
				val := marshalData(payload)
				sendToAll(room.users, val)

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
				payload := map[string]interface{}{
					"type":  "userLeft",
					"id":    user.name,
					"users": a,
				}

				val := marshalData(payload)
				sendToAll(room.users, val)
			default:
				fmt.Println("Unsupported Type")
			}
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
