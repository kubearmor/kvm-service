// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package core

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ct "github.com/kubearmor/KVMService/src/constants"
	etcd "github.com/kubearmor/KVMService/src/etcd"
	kg "github.com/kubearmor/KVMService/src/log"
	tp "github.com/kubearmor/KVMService/src/types"
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

	EnableVirtualMachinePolicy bool

	// Virtual Machine policies and mappers
	VirtualMachineSecurityPolicies     []tp.VirtualMachineSecurityPolicy
	VirtualMachineSecurityPoliciesLock *sync.RWMutex

	MapIdentityToEWName map[uint16]string
	MapEWNameToIdentity map[string]uint16

	MapIdentityToLabel            map[uint16]string
	MapLabelToIdentities          map[string][]uint16
	MapVirtualMachineConnIdentity map[uint16]ClientConn

	// WgOperatorDaemon Handler
	WgOperatorDaemon sync.WaitGroup
}

// NewKVMSOperatorDaemon Function
func NewKVMSOperatorDaemon() *KVMSOperator {
	dm := new(KVMSOperator)

	dm.EtcdClient = etcd.NewEtcdClient()

	dm.VirtualMachineSecurityPoliciesLock = new(sync.RWMutex)

	dm.MapIdentityToLabel = make(map[uint16]string)
	dm.MapLabelToIdentities = make(map[string][]uint16)

	dm.MapIdentityToEWName = make(map[uint16]string)
	dm.MapEWNameToIdentity = make(map[string]uint16)

	dm.MapVirtualMachineConnIdentity = make(map[uint16]ClientConn)

	dm.WgOperatorDaemon = sync.WaitGroup{}
	kg.Print("Successfully initialized KVMSOperator")

	return dm
}

// DestroyKVMSOperator Function
func (dm *KVMSOperator) DestroyKVMSOperator() {

	// wait for a while
	time.Sleep(time.Second * 1)

	// wait for other routines
	kg.Print("Waiting for remaining routine terminations")

	kg.Print("Deleting the Virtual Machine keys from etcd")
	dm.EtcdClient.EtcdDelete(context.TODO(), ct.KvmOprLabelToIdentities)

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
func KVMSOperatorDaemon() {

	// create a daemon
	dm := NewKVMSOperatorDaemon()

	// wait for a while
	time.Sleep(time.Second * 1)

	if K8s.InitK8sClient() {
		kg.Print("Started the Virtual Machine CRD watcher")
		go dm.WatchVirtualMachineSecurityPolicies()

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
