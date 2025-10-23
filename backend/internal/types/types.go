package types

import "time"

type Item struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
