package types

import (
	"github.com/kubearmor/KVMService/src/service/cilium/labels"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Rule struct {
	EndpointSelector *v1.LabelSelector `json:"endpointSelector,omitempty"`
	NodeSelector     *v1.LabelSelector `json:"nodeSelector,omitempty"`
	Ingress          []IngressRule     `json:"ingress,omitempty"`
	IngressDeny      []IngressDenyRule `json:"ingressDeny,omitempty"`
	Egress           []EgressRule      `json:"egress,omitempty"`
	EgressDeny       []EgressDenyRule  `json:"egressDeny,omitempty"`
	Labels           labels.LabelArray `json:"labels,omitempty"`
	Description      string            `json:"description,omitempty"`
	AuditMode        bool              `json:"auditMode,omitempty,default:true"`
}

type IngressRule struct {
	IngressCommonRule `json:",inline"`
	ToPorts           PortRules `json:"toPorts,omitempty"`
}

type IngressDenyRule struct {
	IngressCommonRule `json:",inline"`
	ToPorts           PortDenyRules `json:"toPorts,omitempty"`
}

type EgressRule struct {
	EgressCommonRule `json:",inline"`
	ToPorts          PortRules         `json:"toPorts,omitempty"`
	ToFQDNs          FQDNSelectorSlice `json:"toFQDNs,omitempty"`
}

type EgressDenyRule struct {
	EgressCommonRule `json:",inline"`
	ToPorts          PortDenyRules `json:"toPorts,omitempty"`
}

type IngressCommonRule struct {
	FromEndpoints []v1.LabelSelector `json:"fromEndpoints,omitempty"`
	FromRequires  []v1.LabelSelector `json:"fromRequires,omitempty"`
	FromCIDR      CIDRSlice          `json:"fromCIDR,omitempty"`
	FromCIDRSet   CIDRRuleSlice      `json:"fromCIDRSet,omitempty"`
	FromEntities  EntitySlice        `json:"fromEntities,omitempty"`
}

type EgressCommonRule struct {
	ToEndpoints []v1.LabelSelector `json:"toEndpoints,omitempty"`
	ToRequires  []v1.LabelSelector `json:"toRequires,omitempty"`
	ToCIDR      CIDRSlice          `json:"toCIDR,omitempty"`
	ToCIDRSet   CIDRRuleSlice      `json:"toCIDRSet,omitempty"`
	ToEntities  EntitySlice        `json:"toEntities,omitempty"`
}

type Entity string
type EntitySlice []Entity

type PortRule struct {
	Ports          []PortProtocol `json:"ports,omitempty"`
	TerminatingTLS *TLSContext    `json:"terminatingTLS,omitempty"`
	OriginatingTLS *TLSContext    `json:"originatingTLS,omitempty"`
	Rules          *L7Rules       `json:"rules,omitempty"`
}

type PortRules []PortRule

type PortDenyRule struct {
	Ports []PortProtocol `json:"ports,omitempty"`
}

type PortDenyRules []PortDenyRule

type PortProtocol struct {
	Port     string  `json:"port"`
	Protocol L4Proto `json:"protocol,omitempty"`
}

type L4Proto string

type TLSContext struct {
	Secret      *Secret `json:"secret"`
	TrustedCA   string  `json:"trustedCA,omitempty"`
	Certificate string  `json:"certificate,omitempty"`
	PrivateKey  string  `json:"privateKey,omitempty"`
	Spiffe      *Spiffe `json:"spiffe"`
	DstPort     uint16  `json:"dstPort"`
}

type Secret struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name"`
}

type Spiffe struct {
	PeerIDs []string `json:"peerIDs"`
}

type L7Rules struct {
	HTTP    []PortRuleHTTP  `json:"http,omitempty"`
	Kafka   []KafkaPortRule `json:"kafka,omitempty"`
	DNS     []PortRuleDNS   `json:"dns,omitempty"`
	L7Proto string          `json:"l7proto,omitempty"`
	L7      []PortRuleL7    `json:"l7,omitempty"`
}

type PortRuleHTTP struct {
	Path          string         `json:"path,omitempty"`
	Method        string         `json:"method,omitempty"`
	Host          string         `json:"host,omitempty"`
	Headers       []string       `json:"headers,omitempty"`
	HeaderMatches []*HeaderMatch `json:"headerMatches,omitempty"`
	RuleID        uint16         `json:"ruleID,omitempty"`
	AuditMode     bool           `json:"auditMode,omitempty"`
}

type KafkaPortRule struct {
	Role       string `json:"role,omitempty"`
	APIKey     string `json:"apiKey,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
	ClientID   string `json:"clientID,omitempty"`
	Topic      string `json:"topic,omitempty"`
}

type HeaderMatch struct {
	Mismatch MismatchAction `json:"mismatch,omitempty"`
	Name     string         `json:"name"`
	Secret   *Secret        `json:"secret,omitempty"`
	Value    string         `json:"value,omitempty"`
}

type MismatchAction string

type FQDNSelector struct {
	MatchName    string `json:"matchName,omitempty"`
	MatchPattern string `json:"matchPattern,omitempty"`
}

type PortRuleDNS FQDNSelector

type FQDNSelectorSlice []FQDNSelector

type PortRuleL7 map[string]string
