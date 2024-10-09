package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdClient interface defines the methods we need from etcd client
type EtcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	Close() error
}

// Query wrapper around a query, storing operation name and key
type Query struct {
	Name string
	Key  string
}

// LogFunc defines the signature for the logging function
type LogFunc func(ctx context.Context, q Query, args ...interface{})

type etcdClient struct {
	client  *clientv3.Client
	LogFunc LogFunc
}

// New creates a new EtcdClient with logging
func New(client *clientv3.Client, logFunc LogFunc) EtcdClient {
	return &etcdClient{
		client: client,
		LogFunc: LogFunc(func(ctx context.Context, q Query, args ...interface{}) {
			if logFunc != nil {
				logFunc(ctx, q, args...)
			}
		}),
	}
}

func (e *etcdClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	e.LogFunc(ctx, Query{Name: "Get", Key: key})
	return e.client.Get(ctx, key, opts...)
}

func (e *etcdClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	e.LogFunc(ctx, Query{Name: "Put", Key: key})
	return e.client.Put(ctx, key, val, opts...)
}

func (e *etcdClient) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	e.LogFunc(ctx, Query{Name: "Watch", Key: key})
	return e.client.Watch(ctx, key, opts...)
}

func (e *etcdClient) Close() error {
	return e.client.Close()
}

// DefaultLogFunc provides a default logging implementation
func DefaultLogFunc(ctx context.Context, q Query, args ...interface{}) {
	// Implement your default logging logic here
}
