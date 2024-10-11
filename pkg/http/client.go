package http

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/url"
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
	logger       *zap.Logger
}

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

func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

func NewClient(options ...ClientOption) *Client {
	pc := &Client{
		client: resty.New().
			SetTimeout(15 * time.Second).
			SetRetryCount(0).
			SetLogger(nil),
		logger: zap.NewNop(), // Use a no-op logger by default
	}

	for _, option := range options {
		option(pc)
	}

	return pc
}

func (pc *Client) setProxyForClient(index int) error {
	proxyURL, err := url.Parse(pc.proxies[index].URL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %v", err)
	}
	pc.client.SetProxy(proxyURL.String())
	pc.logger.Info("Set proxy", zap.String("proxy", proxyURL.String()))
	return nil
}

func (pc *Client) Get(url string) (string, error) {
	if len(pc.proxies) > 0 {
		startIndex := pc.currentProxy
		for i := 0; i < len(pc.proxies); i++ {
			currentIndex := (startIndex + i) % len(pc.proxies)

			pc.mu.Lock()
			if !pc.proxies[currentIndex].Working {
				pc.mu.Unlock()
				continue
			}

			err := pc.setProxyForClient(currentIndex)
			if err != nil {
				pc.proxies[currentIndex].Working = false
				pc.proxies[currentIndex].Error = err
				pc.logger.Error("Proxy setup failed", zap.Error(err))
				pc.mu.Unlock()
				continue
			}
			pc.currentProxy = currentIndex
			pc.mu.Unlock()

			pc.logger.Info("Attempting request with proxy", zap.String("proxy", pc.proxies[currentIndex].URL))

			maxRetries := 3
			for retry := 0; retry < maxRetries; retry++ {
				resp, err := pc.client.R().Get(url)
				if err == nil {
					pc.logger.Info("Request successful", zap.String("proxy", pc.proxies[currentIndex].URL))
					return resp.String(), nil
				}

				pc.logger.Warn("Request attempt failed",
					zap.Int("attempt", retry+1),
					zap.String("proxy", pc.proxies[currentIndex].URL),
					zap.Error(err))

				if retry < maxRetries-1 {
					time.Sleep(time.Second * time.Duration(retry+1))
				}
			}

			pc.mu.Lock()
			pc.proxies[currentIndex].Working = false
			pc.proxies[currentIndex].Error = fmt.Errorf("max retries reached")
			pc.mu.Unlock()
		}
		return "", fmt.Errorf("all proxies failed")
	} else {
		pc.logger.Info("No proxies provided, making direct request")
		resp, err := pc.client.R().Get(url)
		if err != nil {
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
