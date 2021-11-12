package constants

var (
	KvmOprIdentityToEWName  = "/kvm-opr-map-identity-to-ewname/"
	KvmOprLabelToIdentities = "/kvm-opr-label-to-identities-map/"
	KvmSvcIdentitiToPodIps  = "/kvm-svc-identity-podip-maps/"
	KvmOprIdentityToLabel   = "/kvm-opr-identity-to-label-maps/"
	KhpCRDName              = "kubearmorhostpolicies"
	LocalHostIPAddress      = "127.0.0.1"
	KubeProxyK8sPort        = "8001"

	EtcdCertFile  = "/etc/kubernetes/pki/etcd/server.crt"
	EtcdKeyFile   = "/etc/kubernetes/pki/etcd/server.key"
	EtcdCAFile    = "/etc/kubernetes/pki/etcd/ca.crt"
	EtcdEndPoints = "https://10.0.2.15:2379"
	EtcdClientTTL = 10
)
