// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
    "context"

	etcd "github.com/kubearmor/KVMService/operator/etcd"
	kg "github.com/kubearmor/KVMService/operator/log"
	tp "github.com/kubearmor/KVMService/operator/types"
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

// KVMSOperator Structure
type KVMSOperator struct {
	EtcdClient *etcd.EtcdClient

	EnableExternalWorkloadPolicy bool

	Port      uint16
	ClusterIp string

	// External workload policies and mappers
	ExternalWorkloadSecurityPolicies     []tp.ExternalWorkloadSecurityPolicy
	ExternalWorkloadSecurityPoliciesLock *sync.RWMutex

	MapIdentityToEWName map[uint16]string
	MapEWNameToIdentity map[string]uint16

	MapIdentityToLabel              map[uint16]string
	MapLabelToIdentity              map[string][]uint16
	MapExternalWorkloadConnIdentity map[uint16]ClientConn

	// WgOperatorDaemon Handler
	WgOperatorDaemon sync.WaitGroup
}

// NewKVMSOperatorDaemon Function
func NewKVMSOperatorDaemon(port int, ipAddress string, enableExternalWorkloadPolicy bool) *KVMSOperator {
	dm := new(KVMSOperator)

	dm.EtcdClient = etcd.NewEtcdClient()

	dm.ClusterIp = ipAddress
	dm.Port = uint16(port)

	dm.EnableExternalWorkloadPolicy = enableExternalWorkloadPolicy
	dm.ExternalWorkloadSecurityPoliciesLock = new(sync.RWMutex)

	dm.MapIdentityToLabel = make(map[uint16]string)
	dm.MapLabelToIdentity = make(map[string][]uint16)

	dm.MapIdentityToEWName = make(map[uint16]string)
	dm.MapEWNameToIdentity = make(map[string]uint16)

	dm.MapExternalWorkloadConnIdentity = make(map[uint16]ClientConn)

	dm.WgOperatorDaemon = sync.WaitGroup{}

	return dm
}

// DestroyKVMSOperator Function
func (dm *KVMSOperator) DestroyKVMSOperator() {

	// wait for a while
	time.Sleep(time.Second * 1)

	// wait for other routines
	kg.Print("Waiting for remaining routine terminations")

	kg.Print("Deleting the external worklaods keys from etcd")
    dm.EtcdClient.EtcdDelete(context.TODO(), "/externalworkloads/")

	dm.WgOperatorDaemon.Wait()
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
func KVMSOperatorDaemon(port int, ipAddress string, enableExternalWorkloadPolicyPtr bool) {
	// create a daemon
	dm := NewKVMSOperatorDaemon(port, ipAddress, enableExternalWorkloadPolicyPtr)

	// wait for a while
	time.Sleep(time.Second * 1)

	// == //

	if K8s.InitK8sClient() {
		if dm.EnableExternalWorkloadPolicy {
			go dm.WatchExternalWorkloadSecurityPolicies()
		}

	} else {
		kg.Print("Kubernetes is not initiliased and Operator is failed!")
	}

	// wait for a while
	time.Sleep(time.Second * 1)

	// listen for interrupt signals
	sigChan := GetOSSigChannel()
	<-sigChan
	close(StopChan)

	// destroy the daemon
	dm.DestroyKVMSOperator()
}
