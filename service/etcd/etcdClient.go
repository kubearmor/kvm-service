// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package etcdClient

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	ct "github.com/kubearmor/KVMService/service/constants"
	kg "github.com/kubearmor/KVMService/service/log"
	tp "github.com/kubearmor/KVMService/service/types"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// var kew_crds []string
// var ew_khps []tp.MK8sKubeArmorHostPolicy

type EtcdClient struct {
	etcdClient    *clientv3.Client
	leaseResponse *clientv3.LeaseGrantResponse
}

func NewEtcdClient() *EtcdClient {
	tlsInfo := transport.TLSInfo{
		CertFile:      ct.EtcdCertFile,
		KeyFile:       ct.EtcdKeyFile,
		TrustedCAFile: ct.EtcdCAFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{ct.EtcdEndPoints},
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	// minimum lease TTL is 5-second
	resp, err := cli.Grant(context.TODO(), int64(ct.EtcdClientTTL))
	if err != nil {
		log.Fatal(err)
	}

	kg.Print("ETCD: etcd client is successfully initialized!!!\n")
	return &EtcdClient{etcdClient: cli, leaseResponse: resp}
}

func (cli *EtcdClient) EtcdPutWithTTL(ctx context.Context, key, value string) error {
	kg.Printf("ETCD: putting values with TTL key:%v value:%v\n", key, value)
	_, err := cli.etcdClient.Put(context.TODO(), key, value, clientv3.WithLease(cli.leaseResponse.ID))
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (cli *EtcdClient) EtcdPut(ctx context.Context, key, value string) error {
	kg.Printf("ETCD: Putting values key:%v value:%v\n", key, value)
	_, err := cli.etcdClient.Put(ctx, key, value)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (cli *EtcdClient) EtcdGetRaw(ctx context.Context, key string) (*clientv3.GetResponse, error) {
	kg.Printf("ETCD: Getting raw values key:%v\n", key)
	resp, err := cli.etcdClient.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		kg.Print("ETCD: err: No data")
		return nil, nil
	}

	return resp, nil
}

func (cli *EtcdClient) EtcdGet(ctx context.Context, key string) (map[string]string, error) {
	kg.Printf("ETCD: Getting values key:%v\n", key)
	keyValuePair := make(map[string]string)
	resp, err := cli.etcdClient.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		kg.Print("ETCD: err: No data")
		return nil, nil
	}

	for _, ev := range resp.Kvs {
		keyValuePair[string(ev.Key)] = string(ev.Value)
		log.Printf("ETCD: Key:%s Value:%s", string(ev.Key), string(ev.Value))
	}

	return keyValuePair, nil
}

func (cli *EtcdClient) EtcdDelete(ctx context.Context, key string) error {
	kg.Printf("ETCD: Deleting key:%v", key)
	_, err := cli.etcdClient.Delete(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (cli *EtcdClient) KeepAliveEtcdConnection() {
	for {
		kg.Print("ETCD: Keep alive etcd connection")
		_, kaerr := cli.etcdClient.KeepAlive(context.TODO(), cli.leaseResponse.ID)
		if kaerr != nil {
			log.Fatal(kaerr)
		}
		time.Sleep(time.Second * 3)
	}
}
