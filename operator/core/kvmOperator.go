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

	cli "github.com/kubearmor/KVMService/operator/clihandler"
	ct "github.com/kubearmor/KVMService/operator/constants"
	etcd "github.com/kubearmor/KVMService/operator/etcd"
	gs "github.com/kubearmor/KVMService/operator/genscript"
	kg "github.com/kubearmor/KVMService/operator/log"
	tp "github.com/kubearmor/KVMService/operator/types"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
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
	CliHandler *cli.Server

	EnableExternalWorkloadPolicy bool

	Port      uint16
	ClusterIp string

	cliPort string
	// External workload policies and mappers
	ExternalWorkloadSecurityPolicies     []tp.ExternalWorkloadSecurityPolicy
	ExternalWorkloadSecurityPoliciesLock *sync.RWMutex

	MapIdentityToEWName map[uint16]string
	MapEWNameToIdentity map[string]uint16

	MapIdentityToLabel              map[uint16]string
	MapLabelToIdentities            map[string][]uint16
	MapExternalWorkloadConnIdentity map[uint16]ClientConn

	// WgOperatorDaemon Handler
	WgOperatorDaemon sync.WaitGroup
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

	kvmService, err := clientset.CoreV1().Services("kube-system").Get(context.Background(), "kvmsoperator", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	for _, lbIngress := range kvmService.Status.LoadBalancer.Ingress {
		externalIp = lbIngress.IP
	}

	kg.Printf("KVMOperator external IP => %v", externalIp)
	return externalIp, nil
}

// NewKVMSOperatorDaemon Function
func NewKVMSOperatorDaemon(port int, ipAddress string) *KVMSOperator {
	dm := new(KVMSOperator)

	dm.EtcdClient = etcd.NewEtcdClient()
	dm.cliPort = ct.KCLIPort
	dm.CliHandler = cli.NewServerInit(dm.cliPort, dm.EtcdClient)

	dm.ClusterIp = ipAddress
	dm.Port = uint16(port)

	dm.ExternalWorkloadSecurityPoliciesLock = new(sync.RWMutex)

	dm.MapIdentityToLabel = make(map[uint16]string)
	dm.MapLabelToIdentities = make(map[string][]uint16)

	dm.MapIdentityToEWName = make(map[uint16]string)
	dm.MapEWNameToIdentity = make(map[string]uint16)

	dm.MapExternalWorkloadConnIdentity = make(map[uint16]ClientConn)

	dm.WgOperatorDaemon = sync.WaitGroup{}
	kg.Printf("Successfully initialized the KVMSOperator with args => (clusterIp:%s clusterPort:%d", dm.ClusterIp, dm.Port)

	return dm
}

// DestroyKVMSOperator Function
func (dm *KVMSOperator) DestroyKVMSOperator() {

	// wait for a while
	time.Sleep(time.Second * 1)

	// wait for other routines
	kg.Print("Waiting for remaining routine terminations")

	kg.Print("Deleting the external worklaods keys from etcd")
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
func KVMSOperatorDaemon(port int, ipAddress string) {
	// Get kvmsoperator external ip
	externalIp, err := getExternalIP()
	if err != nil {
		kg.Err(err.Error())
		return
	}

	// create a daemon
	dm := NewKVMSOperatorDaemon(port, externalIp)

	// wait for a while
	time.Sleep(time.Second * 1)

	// == //
	dm.ClusterIp = externalIp
	gs.InitGenScript(dm.Port, dm.ClusterIp)

	if K8s.InitK8sClient() {
		kg.Print("Started the external workload CRD watcher")
		go dm.WatchExternalWorkloadSecurityPolicies()

		kg.Print("Started the CLI Handler")
		go dm.CliHandler.InitServer()

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
