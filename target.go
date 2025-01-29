package treanteyes

import "github.com/google/uuid"

type Target struct {
	ID          uuid.UUID
	Customer    string
	Name        string
	Description string
}
