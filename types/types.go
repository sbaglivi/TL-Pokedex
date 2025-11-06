package types

import "errors"

type Translation string

const (
	Yoda        Translation = "yoda"
	Shakespeare Translation = "shakespeare"
)

var (
	ErrNotFound = errors.New("not found")
	ErrGeneric  = errors.New("generic error")
)

type Cache interface {
	Get(key string) (any, bool)
	Put(key string, value any)
}
