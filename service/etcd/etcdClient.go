// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

//package etcdClient
package main

import (
	"fmt"
	//"github.com/coreos/etcd/clientv3"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/v3"

	"context"
	"log"
	"os"
	"time"
)

func NewEtcdClient() {
	tlsEnabled := os.Getenv("ETCDCTL_TLS")
	tls := len(tlsEnabled) > 0

	if !tls {
		return
	}

	certFile := os.Getenv("ETCD_CERTFILE")
	keyFile := os.Getenv("ETCD_KEYFILE")
	caFile := os.Getenv("ETCD_CAFILE")
	fmt.Println(caFile, keyFile, certFile)

	if certFile == "" || keyFile == "" || caFile == "" {
		return
	}

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
		Endpoints:   []string{"http://10.0.2.15:2379"},
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("client:", cli)
	fmt.Println("Get keys from etcd")
	//func (c *Client) Get(key string, sort, recursive bool) (*Response, error)
	//resp, err := etcdClient.Get("/registry/security.kubearmor.com/kubearmorexternalworkloads/*", true, true)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = cli.Put(ctx, "sample_key", "sample_value")
	if err != nil {
		log.Fatal(err)
	}
	//resp, err := cli.Get(ctx, "/registry/security.kubearmor.com/kubearmorexternalworkloads", clientv3.WithPrefix(), clientv3.WithKeysOnly())
	resp, err := cli.Get(ctx, "sample_key", clientv3.WithPrefix(), clientv3.WithKeysOnly())
	//	resp, err := etcdClient.Get("/registry/security.kubearmor.com/kubearmorexternalworkloads/external-workload-01", true, true)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Wrong key: ", err.Error())
		return
	}

	fmt.Println(resp)
}

func main() {
	fmt.Println("in main")
	NewEtcdClient()
}
