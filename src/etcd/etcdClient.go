// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package etcdClient

import (
	"context"
	"time"

	kc "github.com/kubearmor/KVMService/src/common"
	ct "github.com/kubearmor/KVMService/src/constants"
	kg "github.com/kubearmor/KVMService/src/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	etcdClient    *clientv3.Client
	leaseResponse *clientv3.LeaseGrantResponse
}

type WatcherAddFunc func(key string, obj interface{})
type WatcherUpdateFunc func(key string, obj interface{})
type WatcherDeleteFunc func(key string)
type WatcherUnmarshalFunc func(bytes []byte) (interface{}, error)

type KeyMeta struct {
	version int64
}

type EtcdWatcher struct {
	client       *EtcdClient
	prefix       string
	cache        map[string]*KeyMeta
	add          WatcherAddFunc
	update       WatcherUpdateFunc
	delete       WatcherDeleteFunc
	unmarshal    WatcherUnmarshalFunc
	stopChan     chan struct{}
	pollInterval time.Duration
}

func NewEtcdClient() *EtcdClient {
	/* TODO : To enable certificates in cluster and validate the same
	 * Works fine with minikube
	tlsInfo := transport.TLSInfo{
		CertFile:      ct.EtcdCertFile,
		KeyFile:       ct.EtcdKeyFile,
		TrustedCAFile: ct.EtcdCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		kg.Err(err.Error())
		return nil
	}
	*/

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{kc.GetEtcdEndPoint(ct.EtcdServiceAccountName)},
		DialTimeout: 5 * time.Second,
		//TLS:         tlsConfig,
	})
	if err != nil {
		kg.Err(err.Error())
		return nil
	}

	// minimum lease TTL is 5-second
	resp, err := cli.Grant(context.TODO(), int64(ct.EtcdClientTTL))
	if err != nil {
		kg.Err(err.Error())
		return nil
	}

	kg.Print("Initialized the ETCD client!")
	return &EtcdClient{etcdClient: cli, leaseResponse: resp}
}

func (cli *EtcdClient) EtcdPutWithTTL(ctx context.Context, key, value string) error {
	kg.Printf("ETCD: putting values with TTL key:%v value:%v\n", key, value)
	_, err := cli.etcdClient.Put(context.TODO(), key, value, clientv3.WithLease(cli.leaseResponse.ID))
	if err != nil {
		kg.Err(err.Error())
		return err
	}
	return nil
}

func (cli *EtcdClient) EtcdPut(ctx context.Context, key, value string) error {
	kg.Printf("ETCD: Putting values key:%v value:%v\n", key, value)
	_, err := cli.etcdClient.Put(ctx, key, value)
	if err != nil {
		kg.Err(err.Error())
		return err
	}
	return nil
}

func (cli *EtcdClient) EtcdGetRaw(ctx context.Context, key string) (*clientv3.GetResponse, error) {
	kg.Printf("ETCD: Getting raw values key:%v\n", key)
	resp, err := cli.etcdClient.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		kg.Err(err.Error())
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		kg.Print("ETCD: err: No data")
	}

	return resp, nil
}

func (cli *EtcdClient) EtcdGet(ctx context.Context, key string) (map[string]string, error) {
	kg.Printf("ETCD: Getting values key:%v\n", key)
	keyValuePair := make(map[string]string)
	resp, err := cli.etcdClient.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		kg.Err(err.Error())
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		kg.Print("ETCD: err: No data")
		return nil, nil
	}

	for _, ev := range resp.Kvs {
		keyValuePair[string(ev.Key)] = string(ev.Value)
		kg.Printf("ETCD: Key:%s Value:%s", string(ev.Key), string(ev.Value))
	}
	return keyValuePair, nil
}

func (cli *EtcdClient) EtcdDelete(ctx context.Context, key string) error {
	kg.Printf("ETCD: Deleting key:%v", key)
	_, err := cli.etcdClient.Delete(ctx, key, clientv3.WithPrefix())
	if err != nil {
		kg.Err(err.Error())
		return err
	}
	return nil
}

func (cli *EtcdClient) KeepAliveEtcdConnection() {
	for {
		_, kaerr := cli.etcdClient.KeepAlive(context.TODO(), cli.leaseResponse.ID)
		if kaerr != nil {
			kg.Err(kaerr.Error())
			kg.Print("ETCD: etcd connection handshake failed")
		}
		time.Sleep(time.Second * 3)
	}
}

func NewWatcher(client *EtcdClient, prefix string, interval time.Duration, unmarshal WatcherUnmarshalFunc,
	add WatcherAddFunc, update WatcherUpdateFunc, delete WatcherDeleteFunc) *EtcdWatcher {
	cache := make(map[string]*KeyMeta)
	stop := make(chan struct{}, 1)

	return &EtcdWatcher{
		client:       client,
		prefix:       prefix,
		cache:        cache,
		add:          add,
		update:       update,
		delete:       delete,
		unmarshal:    unmarshal,
		stopChan:     stop,
		pollInterval: interval,
	}
}

func (w *EtcdWatcher) Observe(ctx context.Context) {
	go func() {
		for {
			newIteration := true
			visited := make(map[string]interface{})

		EtcdGet:
			select {
			case <-w.stopChan:
				return
			case <-w.client.etcdClient.Ctx().Done():
				return
			default:
			}

			resp, err := w.client.EtcdGetRaw(ctx, w.prefix)
			if err != nil {
				kg.Err(err.Error())
				goto Sleep
			}

			if len(resp.Kvs) == 0 {
				if newIteration {
					for k := range w.cache {
						if w.delete != nil {
							w.delete(k)
						}
					}
					goto Sleep
				} else {
					goto DeleteKey
				}
			}

			for _, kv := range resp.Kvs {
				key := string(kv.Key)
				visited[key] = nil

				val, err := w.unmarshal(kv.Value)
				if err != nil {
					kg.Err(err.Error())
				}

				if meta, ok := w.cache[key]; ok {
					// Already existing key
					if meta.version != kv.Version {
						meta.version = kv.Version
						if w.update != nil {
							w.update(key, val)
						}
					}
				} else {
					// New key
					w.cache[key] = &KeyMeta{version: kv.Version}
					if w.add != nil {
						w.add(key, val)
					}
				}
			}

			if resp.More {
				// Do EtcdGet() again.
				// Don't jump to check for deleted keys.
				newIteration = false
				goto EtcdGet
			}

		DeleteKey:
			// Check for deleted keys
			for k := range w.cache {
				if _, ok := visited[k]; !ok {
					delete(w.cache, k)
					if w.delete != nil {
						w.delete(k)
					}
				}
			}

		Sleep:
			time.Sleep(w.pollInterval)
		}
	}()
}

func (w *EtcdWatcher) Stop() {
	close(w.stopChan)
}
