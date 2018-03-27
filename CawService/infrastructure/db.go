package infrastructure

import "errors"

const (
	db = "test"
)

var (
	ErrNotFound   = errors.New("Not Found")
	ErrUserExists = errors.New("User Exists")
)
