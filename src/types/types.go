// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of KubeArmor

package types

import (
	"regexp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============ //
// == Docker == //
// ============ //

// Container Structure
type Container struct {
	ContainerID   string `json:"containerID"`
	ContainerName string `json:"containerName"`

	NamespaceName string `json:"namespaceName"`
	EndPointName  string `json:"endPointName"`

	AppArmorProfile string `json:"apparmorProfile"`

	// == //

	PidNS uint32 `json:"pidns"`
	MntNS uint32 `json:"mntns"`

	// == //

	PolicyEnabled int `json:"policyEnabled"`

	ProcessVisibilityEnabled      bool `json:"processVisibilityEnabled"`
	FileVisibilityEnabled         bool `json:"fileVisibilityEnabled"`
	NetworkVisibilityEnabled      bool `json:"networkVisibilityEnabled"`
	CapabilitiesVisibilityEnabled bool `json:"capabilitiesVisibilityEnabled"`
}

// HostVolumeMount Structure
type HostVolumeMount struct {
	Type                    string
	VolumeName              string
	PathName                string
	UsedByContainerPath     map[string]string // key: container name, val: mounted path
	UsedByContainerReadOnly map[string]bool   // key: container name, val: readOnly
}

// EndPoint Structure
type EndPoint struct {
	NamespaceName string `json:"namespaceName"`
	EndPointName  string `json:"endPointName"`

	Labels     map[string]string `json:"labels"`
	Identities []string          `json:"identities"`

	Containers       []string `json:"containers"`
	AppArmorProfiles []string `json:"apparmorProfiles"`

	SELinuxProfiles map[string]string `json:"selinuxProfiles"`
	HostVolumes     []HostVolumeMount `json:"hostVolumes"`

	SecurityPolicies []SecurityPolicy `json:"securityPolicies"`

	// == //

	PolicyEnabled int `json:"policyEnabled"`

	ProcessVisibilityEnabled      bool `json:"processVisibilityEnabled"`
	FileVisibilityEnabled         bool `json:"fileVisibilityEnabled"`
	NetworkVisibilityEnabled      bool `json:"networkVisibilityEnabled"`
	CapabilitiesVisibilityEnabled bool `json:"capabilitiesVisibilityEnabled"`
}

// Node Structure
type Node struct {
	NodeName string `json:"nodeName"`
	NodeIP   string `json:"nodeIP"`

	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`

	Identities []string `json:"identities"`

	Architecture    string `json:"architecture"`
	OperatingSystem string `json:"operatingSystem"`
	OSImage         string `json:"osImage"`
	KernelVersion   string `json:"kernelVersion"`
	KubeletVersion  string `json:"kubeletVersion"`

	ContainerRuntimeVersion string `json:"containerRuntimeVersion"`

	EnableKubeArmorPolicy     bool `json:"enableKubeArmorPolicy"`
	EnableKubeArmorHostPolicy bool `json:"enableKubeArmorHostPolicy"`
}

// ================ //
// == Kubernetes == //
// ================ //

// K8sNodeEvent Structure
type K8sNodeEvent struct {
	Type   string  `json:"type"`
	Object v1.Node `json:"object"`
}

// K8sPod Structure
type K8sPod struct {
	Metadata    map[string]string
	Annotations map[string]string
	Labels      map[string]string
	Containers  map[string]string
	HostVolumes []HostVolumeMount
}

// K8sPodEvent Structure
type K8sPodEvent struct {
	Type   string `json:"type"`
	Object v1.Pod `json:"object"`
}

// K8sPolicyStatus Structure
type K8sPolicyStatus struct {
	Status string `json:"status,omitempty"`
}

// K8sKubeArmorPolicyEvent Structure
type K8sKubeArmorPolicyEvent struct {
	Type   string             `json:"type"`
	Object K8sKubeArmorPolicy `json:"object"`
}

// K8sKubeArmorPolicy Structure
type K8sKubeArmorPolicy struct {
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     SecuritySpec      `json:"spec"`
	Status   K8sPolicyStatus   `json:"status,omitempty"`
}

// K8sKubeArmorPolicies Structure
type K8sKubeArmorPolicies struct {
	Items []K8sKubeArmorPolicy `json:"items"`
}

// K8sKubeArmorHostPolicyEvent Structure
type K8sKubeArmorHostPolicyEvent struct {
	Type   string                 `json:"type"`
	Object K8sKubeArmorHostPolicy `json:"object"`
}

// K8sKubeArmorHostPolicy Structure
type K8sKubeArmorHostPolicy struct {
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     HostSecuritySpec  `json:"spec"`
	Status   K8sPolicyStatus   `json:"status,omitempty"`
}

type MK8sKubeArmorHostPolicy struct {
	Api      string            `json:"apiVersion"`
	Kind     string            `json:"kind"`
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     HostSecuritySpec  `json:"spec"`
}

// K8sKubeArmorHostPolicies Structure
type K8sKubeArmorHostPolicies struct {
	Items []K8sKubeArmorHostPolicy `json:"items"`
}

type K8sKubeArmorVirtualMachinePolicyStatus struct {
	ID     uint64 `json:"id,omitempty"`
	IP     string `json:"ip,omitempty"`
	Status string `json:"status,omitempty"`
}

// K8sKubeArmorVirtualMachinePolicyEvent Structure
type K8sKubeArmorVirtualMachinePolicyEvent struct {
	Type   string                           `json:"type"`
	Object K8sKubeArmorVirtualMachinePolicy `json:"object"`
}

// K8sKubeArmoVirtualMachinePolicy Structure
type K8sKubeArmorVirtualMachinePolicy struct {
	Metadata metav1.ObjectMeta                      `json:"metadata"`
	Spec     VirtualMachineSecuritySpec             `json:"spec"`
	Status   K8sKubeArmorVirtualMachinePolicyStatus `json:"status,omitempty"`
}

// K8sKubeArmorVirtualMachinePolicies Structure
type K8sKubeArmorVirtualMachinePolicies struct {
	Items []K8sKubeArmorVirtualMachinePolicy `json:"items"`
}

// ============= //
// == Logging == //
// ============= //

// Log Structure
type Log struct {
	// updated time
	Timestamp   int64  `json:"timestamp"`
	UpdatedTime string `json:"updatedTime"`

	// host
	ClusterName string `json:"clusterName,omitempty"`
	HostName    string `json:"hostName"`

	// k8s
	NamespaceName string `json:"namespaceName,omitempty"`
	PodName       string `json:"podName,omitempty"`

	// container
	ContainerID   string `json:"containerID,omitempty"`
	ContainerName string `json:"containerName,omitempty"`

	// common
	HostPID int32 `json:"hostPid"`
	PPID    int32 `json:"ppid"`
	PID     int32 `json:"pid"`
	UID     int32 `json:"uid"`

	// policy
	PolicyName string `json:"policyName,omitempty"`

	// severity, tags, message
	Severity string `json:"severity,omitempty"`
	Tags     string `json:"tags,omitempty"`
	Message  string `json:"message,omitempty"`

	// log
	Type      string `json:"type"`
	Source    string `json:"source"`
	Operation string `json:"operation"`
	Resource  string `json:"resource"`
	Data      string `json:"data,omitempty"`
	Action    string `json:"action,omitempty"`
	Result    string `json:"result"`

	// == //

	PolicyEnabled int `json:"policyEnabled,omitempty"`

	ProcessVisibilityEnabled      bool `json:"processVisibilityEnabled,omitempty"`
	FileVisibilityEnabled         bool `json:"fileVisibilityEnabled,omitempty"`
	NetworkVisibilityEnabled      bool `json:"networkVisibilityEnabled,omitempty"`
	CapabilitiesVisibilityEnabled bool `json:"capabilitiesVisibilityEnabled,omitempty"`
}

// MatchPolicy Structure
type MatchPolicy struct {
	PolicyName string

	Severity string
	Tags     []string
	Message  string

	Source       string
	Operation    string
	ResourceType string
	Resource     string

	Regexp *regexp.Regexp
	Native bool

	Action string
}

// MatchPolicies Structure
type MatchPolicies struct {
	Policies []MatchPolicy
}

// ===================== //
// == Security Policy == //
// ===================== //

// KubeArmorPolicy Flags
const (
	KubeArmorPolicyDisabled = 0
	KubeArmorPolicyEnabled  = 1
	KubeArmorPolicyAudited  = 2
)

// SelectorType Structure
type SelectorType struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	Identities  []string          `json:"identities,omitempty"` // set during policy update
}

// MatchSourceType Structure
type MatchSourceType struct {
	Path      string `json:"path,omitempty"`
	Directory string `json:"dir,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
}

// ProcessPathType Structure
type ProcessPathType struct {
	Path       string            `json:"path"`
	OwnerOnly  bool              `json:"ownerOnly,omitempty"`
	FromSource []MatchSourceType `json:"fromSource,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// ProcessDirectoryType Structure
type ProcessDirectoryType struct {
	Directory  string            `json:"dir"`
	Recursive  bool              `json:"recursive,omitempty"`
	OwnerOnly  bool              `json:"ownerOnly,omitempty"`
	FromSource []MatchSourceType `json:"fromSource,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// ProcessPatternType Structure
type ProcessPatternType struct {
	Pattern   string `json:"pattern"`
	OwnerOnly bool   `json:"ownerOnly,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// ProcessType Structure
type ProcessType struct {
	MatchPaths       []ProcessPathType      `json:"matchPaths,omitempty"`
	MatchDirectories []ProcessDirectoryType `json:"matchDirectories,omitempty"`
	MatchPatterns    []ProcessPatternType   `json:"matchPatterns,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// FilePathType Structure
type FilePathType struct {
	Path       string            `json:"path"`
	ReadOnly   bool              `json:"readOnly,omitempty"`
	OwnerOnly  bool              `json:"ownerOnly,omitempty"`
	FromSource []MatchSourceType `json:"fromSource,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// FileDirectoryType Structure
type FileDirectoryType struct {
	Directory  string            `json:"dir"`
	ReadOnly   bool              `json:"readOnly,omitempty"`
	Recursive  bool              `json:"recursive,omitempty"`
	OwnerOnly  bool              `json:"ownerOnly,omitempty"`
	FromSource []MatchSourceType `json:"fromSource,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// FilePatternType Structure
type FilePatternType struct {
	Pattern   string `json:"pattern"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
	OwnerOnly bool   `json:"ownerOnly,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// FileType Structure
type FileType struct {
	MatchPaths       []FilePathType      `json:"matchPaths,omitempty"`
	MatchDirectories []FileDirectoryType `json:"matchDirectories,omitempty"`
	MatchPatterns    []FilePatternType   `json:"matchPatterns,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// NetworkProtocolType Structure
type NetworkProtocolType struct {
	Protocol   string            `json:"protocol"`
	FromSource []MatchSourceType `json:"fromSource,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// NetworkType Structure
type NetworkType struct {
	MatchProtocols []NetworkProtocolType `json:"matchProtocols,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// CapabilitiesCapabilityType Structure
type CapabilitiesCapabilityType struct {
	Capability string            `json:"capability"`
	FromSource []MatchSourceType `json:"fromSource,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// CapabilitiesType Structure
type CapabilitiesType struct {
	MatchCapabilities []CapabilitiesCapabilityType `json:"matchCapabilities,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// MatchVolumeMountType Structure
type MatchVolumeMountType struct {
	Path      string `json:"path,omitempty"`
	Directory string `json:"dir,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// SELinuxType Structure
type SELinuxType struct {
	MatchVolumeMounts []MatchVolumeMountType `json:"matchVolumeMounts,omitempty"`

	Severity int      `json:"severity,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action,omitempty"`
}

// SecuritySpec Structure
type SecuritySpec struct {
	Selector SelectorType `json:"selector"`

	Process      ProcessType      `json:"process,omitempty"`
	File         FileType         `json:"file,omitempty"`
	Network      NetworkType      `json:"network,omitempty"`
	Capabilities CapabilitiesType `json:"capabilities,omitempty"`

	AppArmor string      `json:"apparmor,omitempty"`
	SELinux  SELinuxType `json:"selinux,omitempty"`

	Severity int      `json:"severity"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action"`
}

// SecurityPolicy Structure
type SecurityPolicy struct {
	Metadata map[string]string `json:"metadata"`
	Spec     SecuritySpec      `json:"spec"`
}

// ========================== //
// == Host Security Policy == //
// ========================== //

// NodeSelectorType Structure
type NodeSelectorType struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
	Identities  []string          `json:"identities,omitempty"` // set during policy update
}

// HostSecuritySpec Structure
type HostSecuritySpec struct {
	NodeSelector NodeSelectorType `json:"nodeSelector"`

	Process      ProcessType      `json:"process,omitempty"`
	File         FileType         `json:"file,omitempty"`
	Network      NetworkType      `json:"network,omitempty"`
	Capabilities CapabilitiesType `json:"capabilities,omitempty"`

	AppArmor string `json:"apparmor,omitempty"`

	Severity int      `json:"severity"`
	Tags     []string `json:"tags,omitempty"`
	Message  string   `json:"message,omitempty"`
	Action   string   `json:"action"`
}

// HostSecurityPolicy Structure
type HostSecurityPolicy struct {
	Metadata map[string]string `json:"metadata"`
	Spec     HostSecuritySpec  `json:"spec"`
}

// VirtualMachineSecuritySpec Structure
type VirtualMachineSecuritySpec struct {
	IPv4AllocCIDR string `json:"ipv4-alloc-cidr,omitempty"`
	IPv6AllocCIDR string `json:"ipv6-alloc-cidr,omitempty"`
}

type VirtualMachineMetadata struct {
	NodeSelector NodeSelectorType `json:"nodeSelector"`
	Name         string           `json:"name"`
}

// VirtualMachineSecurityPolicy Structure
type VirtualMachineSecurityPolicy struct {
	Metadata VirtualMachineMetadata `json:"metadata"`
}

// K8sKubeArmorHostPolicyEventWithIdentities
type K8sKubeArmorHostPolicyEventWithIdentity struct {
	Event           K8sKubeArmorHostPolicyEvent
	Identity        uint16
	CloseConnection bool
}

/*
// ================== //
// == Process Tree == //
// ================== //

// PidMap for host pid -> process node
type PidMap map[uint32]PidNode

// PidNode Structure
type PidNode struct {
	PidID uint32
	MntID uint32

	HostPPID uint32
	HostPID  uint32

	PPID uint32
	PID  uint32
	UID  uint32

	Comm     string
	ExecPath string

	Exited     bool
	ExitedTime time.Time
}
*/
