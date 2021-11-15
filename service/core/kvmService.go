// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	//"context"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	etcd "github.com/kubearmor/KVMService/service/etcd"
	kg "github.com/kubearmor/KVMService/service/log"
	ks "github.com/kubearmor/KVMService/service/server"
	tp "github.com/kubearmor/KVMService/service/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
)

// ClientConn is the wrapper for a grpc client conn
type ClientConn struct {
	// conn      *grpc.ClientConn
	// unhealthy bool
}

// ====================== //
// == KubeArmor Daemon == //
// ====================== //

// StopChan Channel
var StopChan chan struct{}

// init Function
func init() {
	StopChan = make(chan struct{})
}

// KVMS Structure
type KVMS struct {
	EtcdClient *etcd.EtcdClient
	Server     *ks.Server

	// gRPC
	gRPCPort  string
	LogPath   string
	LogFilter string

	IdentityConnPool []ClientConn

	MapEtcdEWIdentityLabels map[string]string
	EtcdEWLabels            []string

	// Host Security policies
	HostSecurityPolicies     []tp.HostSecurityPolicy
	HostSecurityPoliciesLock *sync.RWMutex

	// External workload policies and mappers
	ExternalWorkloadSecurityPolicies     []tp.ExternalWorkloadSecurityPolicy
	ExternalWorkloadSecurityPoliciesLock *sync.RWMutex

	MapIdentityToLabel              map[uint16]string
	MapLabelToIdentities            map[string][]uint16
	MapExternalWorkloadConnIdentity map[uint16]ClientConn

	ClusterPort      uint16
	ClusteripAddress string
	PodIpAddress     string

	// WgDaemon Handler
	WgDaemon sync.WaitGroup
}

func getExternalIP() (string, error) {

	/* Calculated time manually to see that the kvmsoperator service
	 * takes a minimum of 45 seconds to fetch the same.
	 * Hence placing a time delay of 1 minute
	 */
	time.Sleep(1 * 60 * time.Second)

	var externalIp string

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		kg.Err(err.Error())
		return "", err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		kg.Err(err.Error())
		return "", err
	}

	kvmService, err := clientset.CoreV1().Services("").Get(context.Background(), "kvmservice", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, lbIngress := range kvmService.Status.LoadBalancer.Ingress {
		externalIp = lbIngress.IP
		break
	}

	kg.Printf("KVMService external IP => %v", externalIp)
	return externalIp, nil
}

// NewKVMSDaemon Function
func NewKVMSDaemon(port int, ipAddress string) *KVMS {
	kg.Print("Initializing all the KVMS daemon attributes")
	dm := new(KVMS)

	dm.EtcdClient = etcd.NewEtcdClient()

	dm.MapEtcdEWIdentityLabels = make(map[string]string)
	dm.EtcdEWLabels = make([]string, 0)

	dm.gRPCPort = ""
	dm.LogPath = ""
	dm.LogFilter = ""
	dm.IdentityConnPool = nil

	dm.ClusterPort = uint16(port)
	dm.ClusteripAddress = ipAddress
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	dm.PodIpAddress = localAddr.IP.String()
	dm.Server = ks.NewServerInit(dm.PodIpAddress, dm.ClusteripAddress, strconv.FormatUint(uint64(dm.ClusterPort), 10), dm.EtcdClient)

	dm.HostSecurityPolicies = []tp.HostSecurityPolicy{}
	dm.HostSecurityPoliciesLock = new(sync.RWMutex)
	dm.ExternalWorkloadSecurityPoliciesLock = new(sync.RWMutex)

	dm.MapIdentityToLabel = make(map[uint16]string)
	dm.MapLabelToIdentities = make(map[string][]uint16)
	dm.MapExternalWorkloadConnIdentity = make(map[uint16]ClientConn)

	dm.WgDaemon = sync.WaitGroup{}
	kg.Print("KVMService attributes got initialized\n")

	return dm
}

// DestroyKVMS Function
func (dm *KVMS) DestroyKVMS() {

	// wait for a while
	time.Sleep(time.Second * 1)

	// wait for other routines
	kg.Print("Waiting for remaining routine terminations")
	dm.WgDaemon.Wait()
}

// ==================== //
// == Signal Handler == //
// ==================== //

// GetOSSigChannel Function
func GetOSSigChannel() chan os.Signal {
	c := make(chan os.Signal, 1)

	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		os.Interrupt)

	return c
}

// ========== //
// == Main == //
// ========== //

// KVMSDaemon Function
func KVMSDaemon(portPtr int) {

	// Get kvmservice external ip
	externalIp, err := getExternalIP()
	if err != nil {
		kg.Err(err.Error())
		return
	}

	// create a daemon
	dm := NewKVMSDaemon(portPtr, externalIp)

	// wait for a while
	time.Sleep(time.Second * 1)

	// == //

	if K8s.InitK8sClient() {
		// watch host security policies
		kg.Print("K8S Client is successfully initialize")

		kg.Print("Watcher triggered for the host policies")
		go dm.WatchHostSecurityPolicies()

		kg.Print("Triggered the keepalive ETCD client")
		go dm.EtcdClient.KeepAliveEtcdConnection()

		kg.Print("Starting gRPC server")
		go dm.Server.InitServer()

	} else {
		kg.Print("K8S client initialization got failed")
	}

	// wait for a while
	time.Sleep(time.Second * 1)

	// listen for interrupt signals
	sigChan := GetOSSigChannel()
	<-sigChan
	close(StopChan)
	close(ks.PolicyChan)

	// destroy the daemon
	dm.DestroyKVMS()
}
