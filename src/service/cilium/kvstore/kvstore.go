package kvstore

import "path"

var (
	BasePrefix = "cilium"

	IdentityPrefix     = path.Join(BasePrefix, "state", "identities", "v1", "id")
	IPCachePrefix      = path.Join(BasePrefix, "state", "ip", "v1")
	NodePrefix         = path.Join(BasePrefix, "state", "nodes", "v1")
	NodeRegisterPrefix = path.Join(BasePrefix, "state", "noderegister", "v1")
	PolicyPrefix       = path.Join(BasePrefix, "state", "policies", "v1", "ccnp")
)
