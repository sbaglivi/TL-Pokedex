package types

import "errors"

type Translation string

const (
	Yoda        Translation = "yoda"
	Shakespeare Translation = "shakespeare"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrTooManyRequests = errors.New("too many requests")
	ErrGeneric         = errors.New("generic error")
)

type Cache interface {
	Get(key string) (any, bool)
	Put(key string, value any)
}

type HTTPError string

const (
	NotFound            HTTPError = "not found"
	InternalServerError HTTPError = "internal server error"
	Timeout             HTTPError = "request timed out"
)

func (err HTTPError) Wrap() map[string]string {
	return map[string]string{"error": string(err)}
}

type Pokemon struct {
	IsLegendary bool   `json:"is_legendary"`
	Name        string `json:"name"`
	Habitat     string `json:"habitat"`
	Desc        string `json:"desc"`
}

type GetPokemonResult struct {
	Pokemon  *Pokemon `json:"pokemon"`
	Warnings []string `json:"warnings,omitempty"`
}
