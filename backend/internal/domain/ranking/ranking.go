package ranking

import "context"

type Entry struct {
	Position    int    `json:"position"`
	AccountName string `json:"accountName"`
}

type Repository interface {
	List(context.Context) ([]Entry, error)
}
