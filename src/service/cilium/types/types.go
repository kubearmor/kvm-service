package types

import (
	"net"

	"github.com/cilium/cilium/pkg/annotation"
	ipamTypes "github.com/cilium/cilium/pkg/ipam/types"
	ciliumv2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/cilium/cilium/pkg/node/addressing"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type Address struct {
	Type addressing.AddressType
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

type RegisterNode struct {
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

func (n *RegisterNode) GetHostIP() net.IP {
	for _, addr := range n.IPAddresses {
		if addr.Type == addressing.NodeInternalIP {
			return addr.IP
		}
	}

	return nil
}

// ToCiliumNode converts the node to a CiliumNode
func (n *RegisterNode) ToCiliumNode() *ciliumv2.CiliumNode {
	var (
		podCIDRs               []string
		ipAddrs                []ciliumv2.NodeAddress
		healthIPv4, healthIPv6 string
		annotations            = map[string]string{}
	)

	if n.IPv4AllocCIDR != nil {
		podCIDRs = append(podCIDRs, n.IPv4AllocCIDR.String())
	}
	if n.IPv6AllocCIDR != nil {
		podCIDRs = append(podCIDRs, n.IPv6AllocCIDR.String())
	}
	if n.IPv4HealthIP != nil {
		healthIPv4 = n.IPv4HealthIP.String()
	}
	if n.IPv6HealthIP != nil {
		healthIPv6 = n.IPv6HealthIP.String()
	}

	for _, address := range n.IPAddresses {
		ipAddrs = append(ipAddrs, ciliumv2.NodeAddress{
			Type: address.Type,
			IP:   address.IP.String(),
		})
	}

	if n.WireguardPubKey != "" {
		annotations[annotation.WireguardPubKey] = n.WireguardPubKey
	}

	return &ciliumv2.CiliumNode{
		ObjectMeta: v1.ObjectMeta{
			Name:        n.Name,
			Labels:      n.Labels,
			Annotations: annotations,
		},
		Spec: ciliumv2.NodeSpec{
			Addresses: ipAddrs,
			HealthAddressing: ciliumv2.HealthAddressingSpec{
				IPv4: healthIPv4,
				IPv6: healthIPv6,
			},
			Encryption: ciliumv2.EncryptionSpec{
				Key: int(n.EncryptionKey),
			},
			IPAM: ipamTypes.IPAMSpec{
				PodCIDRs: podCIDRs,
			},
			NodeIdentity: uint64(n.NodeIdentity),
		},
	}
}
