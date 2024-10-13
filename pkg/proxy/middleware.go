package proxy

import (
	"github.com/go-resty/resty/v2"
)

// Middleware defines the middleware function
type Middleware func(method, url string, req *resty.Request)

// WithMiddleware добавляет middleware к клиенту
func WithMiddleware(middlewares ...Middleware) ClientOption {
	return func(c Client) {
		if client, ok := c.(interface{ SetMiddlewares([]Middleware) }); ok {
			client.SetMiddlewares(middlewares)
		}
	}
}
