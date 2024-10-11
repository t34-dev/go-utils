package proxy

import (
	"net/url"
	"strings"
)

// HTTPClient defines the interface for HTTP clients
type HTTPClient interface {
	Get(url string) (Response, error)
	Post(url string, body interface{}) (Response, error)
	Put(url string, body interface{}) (Response, error)
	Delete(url string) (Response, error)
	Patch(url string, body interface{}) (Response, error)
	GetProxyStatus() []ProxyStatus
}

// Response defines the interface for HTTP responses
type Response interface {
	StatusCode() int
	Body() []byte
	Header() map[string][]string
}

// ProxyStatus represents the status of a proxy
type ProxyStatus struct {
	URL     string
	Working bool
	Error   error
}

// LogFunc defines the signature for the logging function
type LogFunc func(level, msg, proxy string)

// ClientOption is a function type to set options on the Client
type ClientOption func(HTTPClient)

// WithProxy sets the proxies for the client
func WithProxy(proxies []string) ClientOption {
	return func(c HTTPClient) {
		if client, ok := c.(interface{ SetProxies([]string) }); ok {
			client.SetProxies(proxies)
		}
	}
}

// WithLogFunc sets the logging function for the client
func WithLogFunc(logFunc LogFunc) ClientOption {
	return func(c HTTPClient) {
		if client, ok := c.(interface{ SetLogFunc(LogFunc) }); ok {
			client.SetLogFunc(logFunc)
		}
	}
}

// IsConnectionError проверка - ошибка в соединении?
func IsConnectionError(err error) bool {
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
