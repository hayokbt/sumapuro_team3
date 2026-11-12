package ranking

import "context"

type Entry struct {
	Rating      float64 `json:"rating"`
	Position    int     `json:"position"`
	AccountName string  `json:"accountName"`
}

type Repository interface {
	List(context.Context) ([]Entry, error)
}
