// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	//"context"
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
	"google.golang.org/grpc"
)

// ClientConn is the wrapper for a grpc client conn
type ClientConn struct {
	conn      *grpc.ClientConn
	unhealthy bool
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
	MapLabelToIdentity              map[string][]uint16
	MapExternalWorkloadConnIdentity map[uint16]ClientConn

	port      uint16
	ipAddress string

	// WgDaemon Handler
	WgDaemon sync.WaitGroup
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

	dm.port = uint16(port)
	dm.ipAddress = ipAddress

	dm.HostSecurityPolicies = []tp.HostSecurityPolicy{}
	dm.HostSecurityPoliciesLock = new(sync.RWMutex)
	dm.ExternalWorkloadSecurityPoliciesLock = new(sync.RWMutex)

	dm.MapIdentityToLabel = make(map[uint16]string)
	dm.MapLabelToIdentity = make(map[string][]uint16)
	dm.MapExternalWorkloadConnIdentity = make(map[uint16]ClientConn)

	dm.WgDaemon = sync.WaitGroup{}

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

// KVMSService Function
func KVMSDaemon(portPtr int, ipAddressPtr string) {
	// create a daemon
	dm := NewKVMSDaemon(portPtr, ipAddressPtr)

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
		go ks.InitServer(strconv.Itoa(portPtr))

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
