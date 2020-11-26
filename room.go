package main

import (
	"github.com/google/uuid"
)

type room struct {
	users map[uuid.UUID]*user

	id uuid.UUID

	hostID uuid.UUID

	vidID string

	hostName string

	roomSize int
}

func createRoom(id uuid.UUID, vidID, hostID, hostName string) *room {
	return &room{
		users: make(map[uuid.UUID]*user),
		id:    uuid.New(),
	}
}
