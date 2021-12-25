package constants

import "os"

var (
	KvmOprIdentityToEWName  = "/kvm-opr-map-identity-to-ewname/"
	KvmOprEWNameToIdentity  = "/kvm-opr-map-ewname-to-identity/"
	KvmOprLabelToIdentities = "/kvm-opr-label-to-identities-map/"
	KvmSvcIdentitiToPodIps  = "/kvm-svc-identity-podip-maps/"
	KvmOprIdentityToLabel   = "/kvm-opr-identity-to-label-maps/"

	KhpCRDName         = "kubearmorhostpolicies"
	KvmCRDName         = "kubearmorvirtualmachines"
	LocalHostIPAddress = "127.0.0.1"
	KubeProxyK8sPort   = "8001"
	KCLIPort           = "32770"

	CertFile      = "/etc/kubernetes/pki/etcd/server.crt"
	KeyFile       = "/etc/kubernetes/pki/etcd/server.key"
	CAFile        = "/etc/kubernetes/pki/etcd/ca.crt"
	EtcdEndPoints = "https://10.0.2.15:2379"
	EtcdClientTTL = 10

	EtcdServiceAccountName = "etcd0"
	KvmServiceAccountName  = "kvmservice"
	KvmOperatorAccountName = "kvmsoperator"

	EtcdNamespace       = os.Getenv("ETCD_NAMESPACE")
	KvmServiceNamespace = os.Getenv("KVMSERVICE_NAMESPACE")
)
