package http

import "net/http"

type Middleware interface {
	Handler(http.Handler) http.Handler
}
