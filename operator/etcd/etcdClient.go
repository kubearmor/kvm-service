// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package etcdClient

//package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	ct "github.com/kubearmor/KVMService/operator/constants"
	kg "github.com/kubearmor/KVMService/operator/log"
	tp "github.com/kubearmor/KVMService/operator/types"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/v3"
)

var kew_crds []string
var ew_khps []tp.MK8sKubeArmorHostPolicy

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

	kg.Print("Initialized the ETCD client!")
	return &EtcdClient{etcdClient: cli, leaseResponse: resp}
}

func (cli *EtcdClient) EtcdPutWithTTL(ctx context.Context, key, value string) error {
	kg.Printf("ETCD: putting with TTL key:%v value:%v", key, value)
	_, err := cli.etcdClient.Put(context.TODO(), key, value, clientv3.WithLease(cli.leaseResponse.ID))
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (cli *EtcdClient) EtcdPut(ctx context.Context, key, value string) error {
	kg.Printf("ETCD: putting key:%v value:%v", key, value)
	_, err := cli.etcdClient.Put(ctx, key, value)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (cli *EtcdClient) EtcdGet(ctx context.Context, key string) (map[string]string, error) {
	kg.Printf("ETCD: getting key:%v", key)
	keyValuePair := make(map[string]string)
	resp, err := cli.etcdClient.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		kg.Print("ETCD: No data")
		return nil, nil
	}

	for _, ev := range resp.Kvs {
		keyValuePair[string(ev.Key)] = string(ev.Value)
		kg.Printf("ETCD: getting key:%v value:%v", ev.Key, ev.Value)
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

func (cli *EtcdClient) keepAliveEtcdConnection() {
	kg.Print("ETCD: Keep alive etcd connection")
	_, kaerr := cli.etcdClient.KeepAlive(context.TODO(), cli.leaseResponse.ID)
	if kaerr != nil {
		log.Fatal(kaerr)
	}
}

func tempNewEtcdClient() {
	certFile := "/etc/kubernetes/pki/etcd/server.crt"
	keyFile := "/etc/kubernetes/pki/etcd/server.key"
	caFile := "/etc/kubernetes/pki/etcd/ca.crt"
	endPoints := "https://10.0.2.15:2379"

	tlsInfo := transport.TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: caFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endPoints},
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, "/registry/security.kubearmor.com/kubearmorexternalworkloads", clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		log.Fatal(err)
		fmt.Println("Wrong key: ", err.Error())
		return
	}

	for _, ev := range resp.Kvs {
		kew_crds = append(kew_crds, string(ev.Key))
	}

	hostPolicies, err := cli.Get(ctx, "/registry/security.kubearmor.com/kubearmorhostpolicies", clientv3.WithPrefix())

	for _, hp := range hostPolicies.Kvs {
		event := tp.MK8sKubeArmorHostPolicy{}
		if err = json.Unmarshal([]byte(hp.Value), &event); err != nil {
			panic(err)
		}
		if len(event.Spec.NodeSelector.MatchLabels["kubearmorexternalworkloads.security.kubearmor.com"]) > 0 {
			ew_khps = append(ew_khps, event)
		}

		//for _, label := range event.Spec.NodeSelector.MatchLabels {
		//	fmt.Println(label)
		//}
	}

	for _, hp := range ew_khps {
		fmt.Printf("+%v", hp)
	}

	// minimum lease TTL is 5-second
	log.Print("Getting the lease of 2 seconds")
	l_resp, err := cli.Grant(context.TODO(), 2)
	if err != nil {
		log.Fatal(err)
	}

	// after 5 seconds, the key 'foo' will be removed
	log.Print("Putting 'foo':'bar' into etcd")
	_, err = cli.Put(context.TODO(), "foo", "bar", clientv3.WithLease(l_resp.ID))
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Getting 'foo' from the etcd")
	s, err := cli.Get(context.TODO(), "foo", clientv3.WithPrefix())

	if len(s.Kvs) == 0 {
		fmt.Println("No more 'foo'")
	} else {
		fmt.Println("Found key 'foo'")
	}

	log.Print("Sleeping for 3 seconds")
	time.Sleep(3 * time.Second)

	log.Print("Getting 'foo' from the etcd after lease expiry")
	s, err = cli.Get(context.TODO(), "foo", clientv3.WithPrefix())

	if len(s.Kvs) == 0 {
		fmt.Println("No more 'foo'")
	} else {
		fmt.Println("Found key 'foo'")
	}

}

func mmain() {
	log.Println("Creating new etcd client")
	//etcdClient = NewEtcdClient()
}
