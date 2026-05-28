package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ─── Per-check result structs ────────────────────────────────────────────────

type EtcdCheckResult struct {
	Status         string `json:"status"` // OK / Warning / Critical
	DBSizeMB       int64  `json:"dbSizeMB,omitempty"`
	Leader         string `json:"leader,omitempty"`
	MemberCount    int    `json:"memberCount,omitempty"`
	HealthyMembers int    `json:"healthyMembers,omitempty"`
	Message        string `json:"message,omitempty"`
}

type APIServerCheckResult struct {
	Status       string `json:"status"`
	LatencyP99Ms int64  `json:"latencyP99Ms,omitempty"`
	ErrorRate    string `json:"errorRate,omitempty"`
	Message      string `json:"message,omitempty"`
}

type NodeDetail struct {
	Name   string `json:"name"`
	Reason string `json:"reason,omitempty"`
}

type NodeCheckResult struct {
	Status   string       `json:"status"`
	Total    int          `json:"total,omitempty"`
	Ready    int          `json:"ready,omitempty"`
	NotReady []NodeDetail `json:"notReady,omitempty"`
	Message  string       `json:"message,omitempty"`
}

type DegradedOperator struct {
	Name    string `json:"name"`
	Message string `json:"message,omitempty"`
}

type ClusterOperatorCheckResult struct {
	Status   string             `json:"status"`
	Total    int                `json:"total,omitempty"`
	Degraded []DegradedOperator `json:"degraded,omitempty"`
	Message  string             `json:"message,omitempty"`
}

type MCPDetail struct {
	Name                 string `json:"name"`
	MachineCount         int    `json:"machineCount,omitempty"`
	ReadyMachineCount    int    `json:"readyMachineCount,omitempty"`
	DegradedMachineCount int    `json:"degradedMachineCount,omitempty"`
	Status               string `json:"status,omitempty"`
}

type MCPCheckResult struct {
	Status  string      `json:"status"`
	Pools   []MCPDetail `json:"pools,omitempty"`
	Message string      `json:"message,omitempty"`
}

type UnhealthyPod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Phase     string `json:"phase,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type InfraPodCheckResult struct {
	Status            string         `json:"status"`
	CheckedNamespaces int            `json:"checkedNamespaces,omitempty"`
	TotalPodsChecked  int            `json:"totalPodsChecked,omitempty"`
	UnhealthyPods     []UnhealthyPod `json:"unhealthyPods,omitempty"`
	Message           string         `json:"message,omitempty"`
}

type ArgoCDAppDetail struct {
	Name         string `json:"name"`
	HealthStatus string `json:"healthStatus,omitempty"`
	SyncStatus   string `json:"syncStatus,omitempty"`
	Message      string `json:"message,omitempty"`
}

type ArgoCDCheckResult struct {
	Status    string            `json:"status"`
	TotalApps int               `json:"totalApps,omitempty"`
	Healthy   int               `json:"healthy,omitempty"`
	Degraded  []ArgoCDAppDetail `json:"degraded,omitempty"`
	OutOfSync []ArgoCDAppDetail `json:"outOfSync,omitempty"`
	Message   string            `json:"message,omitempty"`
}

type CustomProjectResult struct {
	Name          string         `json:"name"`
	Namespace     string         `json:"namespace"`
	Status        string         `json:"status"`
	UnhealthyPods []UnhealthyPod `json:"unhealthyPods,omitempty"`
	Message       string         `json:"message,omitempty"`
}

type CustomProjectCheckResult struct {
	Status   string                `json:"status"`
	Projects []CustomProjectResult `json:"projects,omitempty"`
	Message  string                `json:"message,omitempty"`
}

// ─── Aggregated check results ────────────────────────────────────────────────

type CheckResults struct {
	Etcd             *EtcdCheckResult            `json:"etcd,omitempty"`
	APIServer        *APIServerCheckResult       `json:"apiServer,omitempty"`
	Nodes            *NodeCheckResult            `json:"nodes,omitempty"`
	ClusterOperators *ClusterOperatorCheckResult `json:"clusterOperators,omitempty"`
	MCPs             *MCPCheckResult             `json:"mcps,omitempty"`
	InfraPods        *InfraPodCheckResult        `json:"infraPods,omitempty"`
	ArgoCD           *ArgoCDCheckResult          `json:"argocd,omitempty"`
	CustomProjects   *CustomProjectCheckResult   `json:"customProjects,omitempty"`
}

// ─── Main Report Spec / Status ───────────────────────────────────────────────

type ClusterHealthReportSpec struct {
	TriggeredBy string       `json:"triggeredBy"` // name of the ClusterHealthCheck CR
	RunTime     *metav1.Time `json:"runTime,omitempty"`
}

type ClusterHealthReportStatus struct {
	OverallStatus string       `json:"overallStatus,omitempty"` // OK / Warning / Critical
	Checks        CheckResults `json:"checks,omitempty"`
	Summary       string       `json:"summary,omitempty"` // human-readable one-liner
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="TriggeredBy",type=string,JSONPath=`.spec.triggeredBy`
// +kubebuilder:printcolumn:name="OverallStatus",type=string,JSONPath=`.status.overallStatus`
// +kubebuilder:printcolumn:name="RunTime",type=string,JSONPath=`.spec.runTime`

type ClusterHealthReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterHealthReportSpec   `json:"spec,omitempty"`
	Status ClusterHealthReportStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterHealthReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterHealthReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterHealthReport{}, &ClusterHealthReportList{})
}
