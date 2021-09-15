// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	kg "github.com/kubearmor/KVMService/service/log"
	tp "github.com/kubearmor/KVMService/service/types"
	"google.golang.org/grpc"
)

// ClientConn is the wrapper for a grpc client conn
type ClientConn struct {
	*grpc.ClientConn
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
	// gRPC
	gRPCPort  string
	LogPath   string
	LogFilter string

	EnableHostPolicy             bool
	EnableExternalWorkloadPolicy bool

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

	/*
			if clusterName == "" {
				if val, ok := os.LookupEnv("CLUSTER_NAME"); ok {
					dm.ClusterName = val
				} else {
					dm.ClusterName = "Default"
				}
			} else {
				dm.ClusterName = clusterName
			}
		dm.gRPCPort = gRPCPort
		dm.LogPath = logPath
		dm.LogFilter = logFilter

		dm.EnableHostPolicy = enableHostPolicy
		dm.EnableExternalWorkloadPolicy = enableExternalWorkloadPolicy
	*/

	dm.gRPCPort = ""
	dm.LogPath = ""
	dm.LogFilter = ""

	dm.EnableHostPolicy = enableHostPolicy
	dm.EnableExternalWorkloadPolicy = enableExternalWorkloadPolicy

	dm.HostSecurityPolicies = []tp.HostSecurityPolicy{}
	dm.HostSecurityPoliciesLock = new(sync.RWMutex)

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

		if dm.EnableExternalWorkloadPolicy {
			go dm.WatchExternalWorkloadSecurityPolicies()
		}

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
