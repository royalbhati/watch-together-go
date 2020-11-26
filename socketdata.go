package main

import "github.com/google/uuid"

type socketData struct {
	ID       string    `json:"id"`
	VidID    string    `json:"vidId"`
	HostID   uuid.UUID `json:"hostId"`
	JoineeID uuid.UUID `json:"joineeID"`
	RoomID   uuid.UUID `json:"roomId"`
	Type     string    `json:"type"`
	Text     string    `json:"text"`
	HostName string    `json:"hostName"`
	UserName string    `json:"userName"`
	Time     string    `json:"time"`
	Status   int       `json:"status"`
}
