package constants

import "os"

var (
	KvmOprIdentityToEWName  = "/kvm-opr-map-identity-to-ewname/"
	KvmOprEWNameToIdentity  = "/kvm-opr-map-ewname-to-identity/"
	KvmOprLabelToIdentities = "/kvm-opr-label-to-identities-map/"
	KvmSvcIdentitiToPodIps  = "/kvm-svc-identity-podip-maps/"
	KvmOprIdentityToLabel   = "/kvm-opr-identity-to-label-maps/"

	KhpCRDName         = "kubearmorhostpolicies"
	KewCRDName         = "kubearmorexternalworkloads"
	LocalHostIPAddress = "127.0.0.1"
	KubeProxyK8sPort   = "8001"
	KCLIPort           = "32770"

	EtcdClientTTL = 10

	EtcdServiceAccountName = "etcd0"
	KvmServiceAccountName  = "kvmservice"
	KvmOperatorAccountName = "kvmsoperator"

	EtcdNamespace       = os.Getenv("ETCD_NAMESPACE")
	KvmServiceNamespace = os.Getenv("KVMSERVICE_NAMESPACE")

	ServerCertPath = "/var/certs/tls.crt"
	ServerKeyPath  = "/var/certs/tls.key"
	CaCertPath     = "/var/ca-certs/ca.cert"
)
