package http

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
	proxies      []ProxyStatus
	client       *resty.Client
	mu           sync.Mutex
	currentProxy int
	logFunc      LogFunc
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
		client:  resty.New(),
		logFunc: func(level, msg string, fields ...interface{}) {}, // Use a no-op log function by default
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
func (pc *Client) setProxyForClient(index int) error {
	proxyURL, err := url.Parse(pc.proxies[index].URL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}
	pc.client.SetProxy(pc.proxies[index].URL)
	pc.logFunc("info", "Use proxy", "proxy", getHealthyProxy(proxyURL))
	return nil
}

func (pc *Client) Get(url string) (string, error) {
	if pc.client == nil {
		return "", fmt.Errorf("client is not initialized")
	}

	err := pc.setProxyForClient(0)
	if err != nil {
		return "", err
	}
	resp, err := pc.client.R().Get(url)
	fmt.Println("-----------", resp.String(), err)
	if err != nil {
		return "", err
	}

	if len(pc.proxies) > 0 {
		var lastError error
		startIndex := pc.currentProxy
		for i := 0; i < len(pc.proxies); i++ {
			currentIndex := (startIndex + i) % len(pc.proxies)

			pc.mu.Lock()
			if !pc.proxies[currentIndex].Working {
				pc.logFunc("info", "Skipping non-working proxy", "proxy", pc.proxies[currentIndex].URL)
				pc.mu.Unlock()
				continue
			}

			err := pc.setProxyForClient(currentIndex)
			if err != nil {
				pc.proxies[currentIndex].Working = false
				pc.proxies[currentIndex].Error = err
				pc.logFunc("error", "Proxy setup failed", "proxy", pc.proxies[currentIndex].URL, "error", err)
				pc.mu.Unlock()
				continue
			}
			pc.currentProxy = currentIndex
			pc.mu.Unlock()

			pc.logFunc("info", "Attempting request with proxy", "proxy", pc.proxies[currentIndex].URL)

			maxRetries := 3
			for retry := 0; retry < maxRetries; retry++ {
				resp, err := pc.client.R().Get(url)
				if err == nil {
					pc.logFunc("info", "Request successful", "proxy", pc.proxies[currentIndex].URL)
					return resp.String(), nil
				}

				pc.logFunc("warn", "Request attempt failed",
					"attempt", retry+1,
					"proxy", pc.proxies[currentIndex].URL,
					"error", err)

				lastError = err

				if retry < maxRetries-1 {
					time.Sleep(time.Second * time.Duration(retry+1))
				}
			}

			pc.mu.Lock()
			pc.proxies[currentIndex].Working = false
			pc.proxies[currentIndex].Error = fmt.Errorf("max retries reached: %w", lastError)
			pc.logFunc("error", "Proxy marked as non-working", "proxy", pc.proxies[currentIndex].URL, "error", pc.proxies[currentIndex].Error)
			pc.mu.Unlock()
		}
		return "", fmt.Errorf("all proxies failed, last error: %w", lastError)
	} else {
		pc.logFunc("info", "No proxies provided, making direct request")
		resp, err := pc.client.R().Get(url)
		if err != nil {
			pc.logFunc("error", "Direct request failed", "error", err)
			return "", err
		}
		return resp.String(), nil
	}
}
func (pc *Client) GetProxyStatus() []ProxyStatus {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.proxies
}
