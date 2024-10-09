package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client represents the etcd client wrapper
type Client interface {
	EtcdClient() EtcdClient
	Close() error
}

type etcdClientWrapper struct {
	client EtcdClient
}

// NewClient creates a new etcd Client
func NewClient(config clientv3.Config, logFunc LogFunc) (Client, error) {
	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	etcdClient := New(client, logFunc)

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
