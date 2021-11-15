// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package etcdClient

//package main

import (
	"context"
	"log"
	"time"

	ct "github.com/kubearmor/KVMService/operator/constants"
	kg "github.com/kubearmor/KVMService/operator/log"
	tp "github.com/kubearmor/KVMService/operator/types"
	clientv3 "go.etcd.io/etcd/client/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var kew_crds []string
var ew_khps []tp.MK8sKubeArmorHostPolicy

type EtcdClient struct {
	etcdClient    *clientv3.Client
	leaseResponse *clientv3.LeaseGrantResponse
}

func getEtcdEndPoint() string {

	var etcdClusterIP string

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		kg.Err(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		kg.Err(err.Error())
	}

	svcList, err := clientset.CoreV1().Services("").List(context.Background(), metav1.ListOptions{FieldSelector: "metadata.name=" + "etcd0"})
	if err != nil {
		return ""
	}

	for _, svc := range svcList.Items {
		etcdClusterIP = svc.Spec.ClusterIP
		break
	}

	etcdClusterIP = "http://" + etcdClusterIP + ":2379"

	kg.Printf("Establishing connection with etcd service => %v", etcdClusterIP)
	return etcdClusterIP
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
		log.Fatal(err)
	}
	*/

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{getEtcdEndPoint()},
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
