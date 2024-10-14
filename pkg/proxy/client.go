package proxy

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strings"
	"sync"
)

type client struct {
	proxies           []ProxyStatus
	client            *resty.Client
	mu                sync.Mutex
	currentProxyIndex int
	logFunc           LogFunc
	middlewares       []Middleware
}

func NewClient(cli *resty.Client, options ...ClientOption) Client {
	rc := &client{
		client:            cli,
		currentProxyIndex: -1,
		logFunc:           func(level, msg, proxy string) {}, // Use a no-op log function by default
		middlewares:       []Middleware{},
	}

	// Устанавливаем кастомную функцию условия повторной попытки
	rc.client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			return !isConnectionError(err)
		}
		return r.StatusCode() >= 500
	})

	for _, option := range options {
		option(rc)
	}

	rc.client.SetLogger(&customLogger{logFunc: rc.logFunc, client: rc})

	return rc
}

func (c *client) Client() *resty.Client {
	return c.client
}

func (c *client) SetMiddlewares(middlewares []Middleware) {
	c.middlewares = middlewares
}

// GetProxyStatus returns the status of all proxies
func (c *client) GetProxyStatus() []ProxyStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.proxies
}
func (c *client) SetProxies(proxies []string) {
	for _, p := range proxies {
		c.proxies = append(c.proxies, ProxyStatus{URL: p, Working: true})
	}
}

func (c *client) SetLogFunc(logFunc LogFunc) {
	c.logFunc = logFunc
}

func (c *client) R() *resty.Request {
	return c.client.R()
}

func (c *client) Get(ctx context.Context, url string, req *resty.Request) (*resty.Response, error) {
	return c.doRequest(ctx, "GET", url, req)
}

func (c *client) Post(ctx context.Context, url string, req *resty.Request) (*resty.Response, error) {
	return c.doRequest(ctx, "POST", url, req)
}

func (c *client) Put(ctx context.Context, url string, req *resty.Request) (*resty.Response, error) {
	return c.doRequest(ctx, "PUT", url, req)
}

func (c *client) Delete(ctx context.Context, url string, req *resty.Request) (*resty.Response, error) {
	return c.doRequest(ctx, "DELETE", url, req)
}

func (c *client) Patch(ctx context.Context, url string, req *resty.Request) (*resty.Response, error) {
	return c.doRequest(ctx, "PATCH", url, req)
}

func (c *client) doRequest(ctx context.Context, method, url string, req *resty.Request) (*resty.Response, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client is not initialized")
	}
	r := c.R()
	if req != nil {
		r = req
	}
	r.URL = c.client.BaseURL + url
	if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "https") {
		r.URL = url
	}

	// use middleware
	for _, middleware := range c.middlewares {
		middleware(ctx, method, url, r)
	}

	if len(c.proxies) == 0 {
		return c.executeRequest(method, url, r)
	} else {
		for i := 0; i < len(c.proxies); i++ {
			ok := c.setWorkProxy()
			if !ok {
				return nil, fmt.Errorf("all proxies are not working")
			}
			resp, err := c.executeRequest(method, url, r)
			if err != nil {
				// Mark current proxy as not working
				c.mu.Lock()
				c.proxies[c.currentProxyIndex].Working = false
				c.proxies[c.currentProxyIndex].Error = err
				c.mu.Unlock()
				c.logFunc("error", fmt.Sprintf("Request failed: %v", err), MaskProxyPassword(c.proxies[c.currentProxyIndex].URL))
				continue
			}
			c.logFunc("info", fmt.Sprintf("Request successful: %s %s", method, url), MaskProxyPassword(c.proxies[c.currentProxyIndex].URL))
			return resp, nil
		}
		return nil, fmt.Errorf("all proxies failed")
	}
}
func (c *client) executeRequest(method, url string, req *resty.Request) (*resty.Response, error) {
	switch method {
	case "GET":
		return req.Get(url)
	case "POST":
		return req.Post(url)
	case "PUT":
		return req.Put(url)
	case "DELETE":
		return req.Delete(url)
	case "PATCH":
		return req.Patch(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
}

func (c *client) getCurrentProxy() string {
	if c.currentProxyIndex >= 0 && c.currentProxyIndex < len(c.proxies) {
		return MaskProxyPassword(c.proxies[c.currentProxyIndex].URL)
	}
	return "No proxy"
}

func (c *client) findWorkProxy() int {
	startIndex := c.currentProxyIndex + 1
	if len(c.proxies) <= startIndex {
		return -1
	}

	for i := startIndex; i < len(c.proxies); i++ {
		if c.proxies[i].Working {
			return i
		}
	}

	return -1
}

func (c *client) setWorkProxy() bool {
	if len(c.proxies) == 0 {
		return false
	}
	idx := c.findWorkProxy()
	if idx == -1 {
		return false
	}
	c.currentProxyIndex = idx
	c.client.SetProxy(c.proxies[c.currentProxyIndex].URL)
	return true
}
