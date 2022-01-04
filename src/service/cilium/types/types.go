package types

import (
	"net"
)

type NumericIdentity uint32

type NamedPort struct {
	Name     string `json:"Name"`
	Port     uint16 `json:"Port"`
	Protocol string `json:"Protocol"`
}

type IPIdentityPair struct {
	IP           net.IP          `json:"IP"`
	Mask         net.IPMask      `json:"Mask"`
	HostIP       net.IP          `json:"HostIP"`
	ID           NumericIdentity `json:"ID"`
	Key          uint8           `json:"Key"`
	Metadata     string          `json:"Metadata"`
	K8sNamespace string          `json:"K8sNamespace,omitempty"`
	K8sPodName   string          `json:"K8sPodName,omitempty"`
	NamedPorts   []NamedPort     `json:"NamedPorts,omitempty"`
}

type AddressType string

const (
	NodeHostName         AddressType = "Hostname"
	NodeExternalIP       AddressType = "ExternalIP"
	NodeInternalIP       AddressType = "InternalIP"
	NodeExternalDNS      AddressType = "ExternalDNS"
	NodeInternalDNS      AddressType = "InternalDNS"
	NodeCiliumInternalIP AddressType = "CiliumInternalIP"
)

type Address struct {
	Type AddressType
	IP   net.IP
}

type Source string

const (
	Unspec         Source = "unspec"
	Local          Source = "local"
	KVStore        Source = "kvstore"
	Kubernetes     Source = "k8s"
	CustomResource Source = "custom-resource"
	Generated      Source = "generated"
)

type Node struct {
	Name            string
	Cluster         string
	IPAddresses     []Address
	IPv4AllocCIDR   *CIDR
	IPv6AllocCIDR   *CIDR
	IPv4HealthIP    net.IP
	IPv6HealthIP    net.IP
	ClusterID       int
	Source          Source
	EncryptionKey   uint8
	Labels          map[string]string
	NodeIdentity    uint32
	WireguardPubKey string
}

const (
	ResourceTypeCiliumNetworkPolicy            = "CiliumNetworkPolicy"
	ResourceTypeCiliumClusterwideNetworkPolicy = "CiliumClusterwideNetworkPolicy"
)

func (n *Node) GetHostIP() net.IP {
	for _, addr := range n.IPAddresses {
		if addr.Type == NodeInternalIP {
			return addr.IP
		}
	}

	return nil
}
