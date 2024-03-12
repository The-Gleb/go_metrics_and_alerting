package repository

import (
	"errors"
)

var (
	ErrConnection          error = errors.New("failed to connect to db")
	ErrNotFound            error = errors.New("metric name not found")
	ErrInvalidMetricStruct       = errors.New("invalid metric struct, some fields are empty, but they shouldn`t")
)
