package pkg

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ProxyStatus struct {
	URL     string
	Working bool
	Error   error
}

type Client struct {
	proxies           []ProxyStatus
	client            *resty.Client
	mu                sync.Mutex
	currentProxyIndex int
	logFunc           LogFunc
}

// LogFunc defines the signature for the logging function
type LogFunc func(level string, msg string, fields ...interface{})

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.SetTimeout(timeout)
	}
}

func WithRetryCount(count int) ClientOption {
	return func(c *Client) {
		c.client.SetRetryCount(count)
	}
}

func WithRetryWaitTime(waitTime, maxWaitTime time.Duration) ClientOption {
	return func(c *Client) {
		c.client.SetRetryWaitTime(waitTime).SetRetryMaxWaitTime(maxWaitTime)
	}
}

func WithProxy(proxies []string) ClientOption {
	return func(c *Client) {
		for _, p := range proxies {
			c.proxies = append(c.proxies, ProxyStatus{URL: p, Working: true})
		}
	}
}

func WithLogFunc(logFunc LogFunc) ClientOption {
	return func(c *Client) {
		c.logFunc = logFunc
		c.client.SetLogger(&customLogger{logFunc: logFunc})
	}
}

func NewClient(options ...ClientOption) *Client {
	pc := &Client{
		currentProxyIndex: -1,
		client:            resty.New(),
		logFunc:           func(level, msg string, fields ...interface{}) {}, // Use a no-op log function by default
	}
	logger := &customLogger{
		logFunc:    pc.logFunc,
		logRetries: true, // По умолчанию логируем повторные запросы
	}
	pc.client.SetLogger(logger)

	// Устанавливаем кастомную функцию условия повторной попытки
	pc.client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Authentication Required") {
				return false // Прекращаем попытки при ошибке аутентификации прокси
			}
			return true // Повторяем попытку для других ошибок
		}
		return r.StatusCode() >= 500 // Повторяем попытку для серверных ошибок
	})

	for _, option := range options {
		option(pc)
	}

	return pc
}

// get healthy proxy address
func getHealthyProxy(addr *url.URL) string {
	u := addr.Scheme + "://"
	if addr.User != nil {
		if addr.User.Username() != "" {
			u += addr.User.Username()
		}
		if _, ok := addr.User.Password(); ok {
			u += ":PW"
		}
		u += "@"
	}
	u += addr.Host + ":" + addr.Port()
	return u
}
func (pc *Client) findWorkProxy() int {
	startIndex := pc.currentProxyIndex + 1
	if len(pc.proxies) <= startIndex {
		return -1 // возвращаем -1, если массив короче 5 элементов
	}

	for i := startIndex; i < len(pc.proxies); i++ {
		if pc.proxies[i].Working {
			return i // возвращаем индекс первого найденного объекта с OK = true
		}
	}

	return -1 // возвращаем -1, если не найдено объекта с OK = true
}
func (pc *Client) setWorkProxy() bool {
	if len(pc.proxies) == 0 {
		return false
	}
	idx := pc.findWorkProxy()
	if idx == -1 {
		return false
	}
	pc.currentProxyIndex = idx
	return true
}
func (pc *Client) setProxyForClient(index int) error {
	proxyURL, err := url.Parse(pc.proxies[index].URL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}
	pc.client.SetProxy(pc.proxies[index].URL)
	pc.logFunc("info", "Use proxy", "proxy", getHealthyProxy(proxyURL))
	return nil
}

func (pc *Client) Get(url string) (*resty.Response, error) {
	if pc.client == nil {
		return nil, fmt.Errorf("client is not initialized")
	}

	if len(pc.proxies) == 0 {
		return pc.client.R().Get(url)
	} else {
		for {
			ok := pc.setWorkProxy()
			if !ok {
				return nil, fmt.Errorf("all proxies are not working")
			}
			resp, err := pc.client.R().Get(url)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Authentication Required") {
					continue
				}
			}
			return resp, nil
		}
	}
}
func (pc *Client) GetProxyStatus() []ProxyStatus {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.proxies
}
