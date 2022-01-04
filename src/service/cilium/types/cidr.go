package types

import "net"

type CIDR struct {
	*net.IPNet
}

type CIDRSlice []string

type CIDRRule struct {
	Cidr        CIDR   `json:"cidr"`
	ExceptCIDRs []CIDR `json:"except,omitempty"`
}
type CIDRRuleSlice []CIDRRule
