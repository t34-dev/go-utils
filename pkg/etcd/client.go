package etcd

import (
	"context"
	"go.etcd.io/etcd/client/v3"
)

// Client represents the etcd client wrapper
type Client interface {
	EtcdClient() EtcdClient
	Close() error
}

type etcdClientWrapper struct {
	client EtcdClient
}

type ClientOption func(*clientOptions)

type clientOptions struct {
	logFunc LogFunc
}

func WithLogFunc(logFunc LogFunc) ClientOption {
	return func(o *clientOptions) {
		o.logFunc = logFunc
	}
}

// NewClient creates a new etcd Client
func NewClient(config clientv3.Config, opts ...ClientOption) (Client, error) {
	options := &clientOptions{
		logFunc: func(ctx context.Context, q Query) {}, // Use a no-op log function by default
	}

	for _, opt := range opts {
		opt(options)
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	etcdClient := New(client, options.logFunc)

	return &etcdClientWrapper{
		client: etcdClient,
	}, nil
}

func (c *etcdClientWrapper) EtcdClient() EtcdClient {
	return c.client
}

func (c *etcdClientWrapper) Close() error {
	return c.client.Close()
}
