package main

import (
	"github.com/google/uuid"
	"github.com/matryer/goblueprints/chapter1/trace"
)

type room struct {
	users map[uuid.UUID]*user

	id uuid.UUID

	hostID uuid.UUID

	vidID string

	hostName string

	roomSize int

	// tracer will receive trace information of activity
	// in the room.
	tracer trace.Tracer
}

// newRoom makes a new room that is ready to
// go.
// const { id, hostId, vidId, hostName, roomSize } = data;

func createRoom(id uuid.UUID, vidID, hostID, hostName string) *room {
	return &room{
		users:  make(map[uuid.UUID]*user),
		tracer: trace.Off(),
		id:     uuid.New(),
	}
}
