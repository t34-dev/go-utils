package proxy

import (
	"context"
	"github.com/go-resty/resty/v2"
)

// Middleware defines the middleware function
type Middleware func(ctx context.Context, method, url string, req *resty.Request)

// WithMiddleware добавляет middleware к клиенту
func WithMiddleware(middlewares ...Middleware) ClientOption {
	return func(c Client) {
		if client, ok := c.(interface{ SetMiddlewares([]Middleware) }); ok {
			client.SetMiddlewares(middlewares)
		}
	}
}
