package entities

import "github.com/google/uuid"

type Status int32

const (
	StatusPending Status = iota
	StatusRunning
	StatusDone
	StatusFailed
)

type Job struct {
	ID      uuid.UUID
	Meta    Meta
	Summary string
	Error   string
	Status  Status
}

type Meta struct {
	Title string
	Text  string
}
