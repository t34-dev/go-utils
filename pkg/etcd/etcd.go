package etcd

import (
	"context"
	"go.etcd.io/etcd/client/v3"
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
	Val  string
}

// LogFunc defines the signature for the logging function
type LogFunc func(ctx context.Context, q Query)

type etcdClient struct {
	client  *clientv3.Client
	logFunc LogFunc
}

// New creates a new EtcdClient with logging
func New(client *clientv3.Client, logFunc LogFunc) EtcdClient {
	if logFunc == nil {
		logFunc = func(ctx context.Context, q Query) {}
	}
	return &etcdClient{
		client:  client,
		logFunc: logFunc,
	}
}

func (e *etcdClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	e.logFunc(ctx, Query{Name: "Get", Key: key})
	return e.client.Get(ctx, key, opts...)
}

func (e *etcdClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	e.logFunc(ctx, Query{Name: "Put", Key: key, Val: val})
	return e.client.Put(ctx, key, val, opts...)
}

func (e *etcdClient) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	e.logFunc(ctx, Query{Name: "Watch", Key: key})
	return e.client.Watch(ctx, key, opts...)
}

func (e *etcdClient) Close() error {
	e.logFunc(context.Background(), Query{Name: "Close"})
	return e.client.Close()
}
