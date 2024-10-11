package proxy_resty

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/t34-dev/go-utils/pkg/proxy"
	"sync"
	"time"
)

type Client struct {
	proxies           []proxy.ProxyStatus
	client            *resty.Client
	mu                sync.Mutex
	currentProxyIndex int
	logFunc           proxy.LogFunc
}

func NewRestyClient(options ...proxy.ClientOption) proxy.HTTPClient {
	rc := &Client{
		client:            resty.New(),
		currentProxyIndex: -1,
		logFunc:           func(level, msg, proxy string) {}, // Use a no-op log function by default
	}

	// Устанавливаем кастомную функцию условия повторной попытки
	rc.client.AddRetryCondition(func(r *resty.Response, err error) bool {
		if err != nil {
			return !proxy.IsConnectionError(err)
		}
		return r.StatusCode() >= 500
	})

	for _, option := range options {
		option(rc)
	}

	return rc
}

// GetProxyStatus returns the status of all proxies
func (c *Client) GetProxyStatus() []proxy.ProxyStatus {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.proxies
}
func (c *Client) SetProxies(proxies []string) {
	for _, p := range proxies {
		c.proxies = append(c.proxies, proxy.ProxyStatus{URL: p, Working: true})
	}
}

func (c *Client) SetLogFunc(logFunc proxy.LogFunc) {
	c.logFunc = logFunc
	c.client.SetLogger(&customLogger{logFunc: logFunc, client: c})
}

func (c *Client) Get(url string) (proxy.Response, error) {
	return c.doRequest("GET", url, nil)
}

func (c *Client) Post(url string, body interface{}) (proxy.Response, error) {
	return c.doRequest("POST", url, body)
}

func (c *Client) Put(url string, body interface{}) (proxy.Response, error) {
	return c.doRequest("PUT", url, body)
}

func (c *Client) Delete(url string) (proxy.Response, error) {
	return c.doRequest("DELETE", url, nil)
}

func (c *Client) Patch(url string, body interface{}) (proxy.Response, error) {
	return c.doRequest("PATCH", url, body)
}

func (c *Client) doRequest(method, url string, body interface{}) (proxy.Response, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client is not initialized")
	}

	if len(c.proxies) == 0 {
		return c.executeRequest(method, url, body)
	} else {
		for i := 0; i < len(c.proxies); i++ {
			ok := c.setWorkProxy()
			if !ok {
				return nil, fmt.Errorf("all proxies are not working")
			}
			resp, err := c.executeRequest(method, url, body)
			if err != nil {
				// Mark current proxy as not working
				c.mu.Lock()
				c.proxies[c.currentProxyIndex].Working = false
				c.proxies[c.currentProxyIndex].Error = err
				c.mu.Unlock()
				c.logFunc("error", fmt.Sprintf("Request failed: %v", err), proxy.MaskProxyPassword(c.proxies[c.currentProxyIndex].URL))
				continue
			}
			c.logFunc("info", fmt.Sprintf("Request successful: %s %s", method, url), proxy.MaskProxyPassword(c.proxies[c.currentProxyIndex].URL))
			return resp, nil
		}
		return nil, fmt.Errorf("all proxies failed")
	}
}

func (c *Client) getCurrentProxy() string {
	if c.currentProxyIndex >= 0 && c.currentProxyIndex < len(c.proxies) {
		return proxy.MaskProxyPassword(c.proxies[c.currentProxyIndex].URL)
	}
	return "No proxy"
}
func (c *Client) executeRequest(method, url string, body interface{}) (proxy.Response, error) {
	request := c.client.R()
	if body != nil {
		request.SetBody(body)
	}

	var resp *resty.Response
	var err error

	switch method {
	case "GET":
		resp, err = request.Get(url)
	case "POST":
		resp, err = request.Post(url)
	case "PUT":
		resp, err = request.Put(url)
	case "DELETE":
		resp, err = request.Delete(url)
	case "PATCH":
		resp, err = request.Patch(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	return &restyResponse{resp}, nil
}

func (c *Client) findWorkProxy() int {
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

func (c *Client) setWorkProxy() bool {
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

type restyResponse struct {
	*resty.Response
}

func (rr *restyResponse) StatusCode() int {
	return rr.Response.StatusCode()
}

func (rr *restyResponse) Body() []byte {
	return rr.Response.Body()
}

func (rr *restyResponse) Header() map[string][]string {
	return rr.Response.Header()
}

// Additional ClientOption functions

func WithTimeout(timeout time.Duration) proxy.ClientOption {
	return func(c proxy.HTTPClient) {
		if client, ok := c.(*Client); ok {
			client.client.SetTimeout(timeout)
		}
	}
}

func WithRetryCount(count int) proxy.ClientOption {
	return func(c proxy.HTTPClient) {
		if client, ok := c.(*Client); ok {
			client.client.SetRetryCount(count)
		}
	}
}

func WithRetryWaitTime(waitTime, maxWaitTime time.Duration) proxy.ClientOption {
	return func(c proxy.HTTPClient) {
		if client, ok := c.(*Client); ok {
			client.client.SetRetryWaitTime(waitTime).SetRetryMaxWaitTime(maxWaitTime)
		}
	}
}
