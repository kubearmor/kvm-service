// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"context"
    "log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	etcd "github.com/kubearmor/KVMService/service/etcd"
	kg "github.com/kubearmor/KVMService/service/log"
	tp "github.com/kubearmor/KVMService/service/types"
	"google.golang.org/grpc"
)

// ClientConn is the wrapper for a grpc client conn
type ClientConn struct {
	conn *grpc.ClientConn
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

	EnableHostPolicy             bool
	EnableExternalWorkloadPolicy bool

    MapEtcdEWIdentityLabels map[string]string
    EtcdEWLabels []string

	// Host Security policies
	HostSecurityPolicies     []tp.HostSecurityPolicy
	HostSecurityPoliciesLock *sync.RWMutex

	// External workload policies and mappers
	ExternalWorkloadSecurityPolicies     []tp.ExternalWorkloadSecurityPolicy
	ExternalWorkloadSecurityPoliciesLock *sync.RWMutex

	MapIdentityToLabel              map[uint16]string
	MapLabelToIdentity              map[string][]uint16
	MapExternalWorkloadConnIdentity map[uint16]ClientConn

	// WgDaemon Handler
	WgDaemon sync.WaitGroup
}

// NewKVMSDaemon Function
func NewKVMSDaemon(enableHostPolicy, enableExternalWorkloadPolicy bool) *KVMS {
	dm := new(KVMS)

	dm.EtcdClient = etcd.NewEtcdClient()
    dm.MapEtcdEWIdentityLabels = make(map[string]string)
    dm.EtcdEWLabels = make([]string, 0)

	dm.gRPCPort = ""
	dm.LogPath = ""
	dm.LogFilter = ""
    dm.IdentityConnPool = nil

    err := dm.EtcdClient.EtcdPut(context.TODO(), "/externalworkloads/34", "abc=yzx")
	if err != nil {
		log.Fatal(err)
	}
	err = dm.EtcdClient.EtcdPut(context.TODO(), "/externalworkloads/12", "abc=yzx")
	if err != nil {
		log.Fatal(err)
	}
	err = dm.EtcdClient.EtcdPut(context.TODO(), "/externalworkloads/foo2", "1235")
	if err != nil {
		log.Fatal(err)
	}
	err = dm.EtcdClient.EtcdPut(context.TODO(), "/externalworkloads/foo3", "1234")
	if err != nil {
		log.Fatal(err)
	}

	dm.EnableHostPolicy = enableHostPolicy
	dm.EnableExternalWorkloadPolicy = enableExternalWorkloadPolicy

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

	// close log feeder
	kg.Print("Stopped the log feeder")

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
func KVMSDaemon(enableExternalWorkloadPolicyPtr, enableHostPolicyPtr bool) {
	// create a daemon
	dm := NewKVMSDaemon(enableExternalWorkloadPolicyPtr, enableHostPolicyPtr)

	// wait for a while
	time.Sleep(time.Second * 1)

	// == //

	if K8s.InitK8sClient() {

		if dm.EnableHostPolicy {
			// watch host security policies
			go dm.WatchHostSecurityPolicies()
		}

		/*
			if dm.EnableExternalWorkloadPolicy {
				go dm.WatchExternalWorkloadSecurityPolicies()
			}*/

	} else {
		kg.Print("dm.EnableExternalWorkloadPolicy true/false")
	}

	// wait for a while
	time.Sleep(time.Second * 1)

	// listen for interrupt signals
	sigChan := GetOSSigChannel()
	<-sigChan
	close(StopChan)

	// destroy the daemon
	dm.DestroyKVMS()
}
