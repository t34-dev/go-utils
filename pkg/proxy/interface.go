package proxy

import (
	"context"
	"github.com/go-resty/resty/v2"
	"net/url"
	"strings"
)

// Client defines the interface for HTTP clients
type Client interface {
	R() *resty.Request
	Client() *resty.Client
	Get(ctx context.Context, url string, req *resty.Request) (*resty.Response, error)
	Post(ctx context.Context, url string, req *resty.Request) (*resty.Response, error)
	Put(ctx context.Context, url string, req *resty.Request) (*resty.Response, error)
	Delete(ctx context.Context, url string, req *resty.Request) (*resty.Response, error)
	Patch(ctx context.Context, url string, req *resty.Request) (*resty.Response, error)
	GetProxyStatus() []ProxyStatus
}

// ProxyStatus represents the status of a proxy
type ProxyStatus struct {
	URL     string
	Working bool
	Error   error
}

// LogFunc defines the signature for the logging function
type LogFunc func(level, msg, proxy string)

// ClientOption is a function type to set options on the client
type ClientOption func(Client)

// WithProxy sets the proxies for the client
func WithProxy(proxies []string) ClientOption {
	return func(c Client) {
		if client, ok := c.(interface{ SetProxies([]string) }); ok {
			client.SetProxies(proxies)
		}
	}
}

// WithLogFunc sets the logging function for the client
func WithLogFunc(logFunc LogFunc) ClientOption {
	return func(c Client) {
		if client, ok := c.(interface{ SetLogFunc(LogFunc) }); ok {
			client.SetLogFunc(logFunc)
		}
	}
}

// IsConnectionError проверка - ошибка в соединении?
func isConnectionError(err error) bool {
	if strings.Contains(err.Error(), "Proxy Authentication Required") {
		return true
	}
	if strings.Contains(err.Error(), "socks connect") {
		return true
	}
	return false
}

// MaskProxyPassword заменяет пароль в URL прокси на звездочки
func MaskProxyPassword(proxyURL string) string {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return proxyURL // Возвращаем исходный URL, если его не удалось разобрать
	}

	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		if password, ok := parsedURL.User.Password(); ok {
			maskedPassword := strings.Repeat("*", 5)
			// Создаем новый URL без экранирования специальных символов
			return strings.Replace(proxyURL, username+":"+password, username+":"+maskedPassword, 1)
		}
	}

	return proxyURL
}
