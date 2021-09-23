// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

//package etcdClient
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	tp "github.com/kubearmor/KVMService/service/types"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	"go.etcd.io/etcd/client/v3"
)

var kew_crds []string
var ew_khps []tp.MK8sKubeArmorHostPolicy

func NewEtcdClient() {
	tlsEnabled := os.Getenv("ETCDCTL_TLS")
	tls := len(tlsEnabled) > 0

	if !tls {
		log.Fatal("TLS is not enabled exiting!!!")
		return
	}

	certFile := os.Getenv("ETCD_CERTFILE")
	keyFile := os.Getenv("ETCD_KEYFILE")
	caFile := os.Getenv("ETCD_CAFILE")

	if certFile == "" || keyFile == "" || caFile == "" {
		log.Fatal("Certs are not configured exiting!!!")
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
		Endpoints:   []string{"https://10.0.2.15:2379"},
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
		fmt.Printf("+%v\n", hp)
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

func main() {
	log.Println("Creating new etcd client")
	NewEtcdClient()
}
