package repository

import (
	"errors"
)

var (
	ErrConnection error = errors.New("failed to connect to db")
	ErrNotFound   error = errors.New("metric name not found")
)
