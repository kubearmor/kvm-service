// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	//"context"

	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	kc "github.com/kubearmor/KVMService/src/common"
	ct "github.com/kubearmor/KVMService/src/constants"
	etcd "github.com/kubearmor/KVMService/src/etcd"
	kg "github.com/kubearmor/KVMService/src/log"
	gs "github.com/kubearmor/KVMService/src/service/genscript"
	ks "github.com/kubearmor/KVMService/src/service/server"
	tp "github.com/kubearmor/KVMService/src/types"
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
	externalIp, err := kc.GetExternalIP(ct.KvmServiceAccountName)
	if err != nil {
		kg.Err(err.Error())
		return
	}

	// create a daemon
	dm := NewKVMSDaemon(portPtr, externalIp)

	// wait for a while
	time.Sleep(time.Second * 1)

	// == //
	gs.InitGenScript(dm.ClusterPort, dm.ClusteripAddress)

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
