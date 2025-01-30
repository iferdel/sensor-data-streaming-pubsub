package treanteyes

import "github.com/google/uuid"

type Asset struct {
	ID          uuid.UUID
	Customer    string
	Name        string
	Description string
}
