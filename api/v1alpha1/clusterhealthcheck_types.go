package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ─── Sub-specs for each check domain ───────────────────────────────────────

type EtcdCheckSpec struct {
	Enabled            bool `json:"enabled"`
	DBSizeLimitMB      int  `json:"dbSizeLimitMB,omitempty"`      // alert if etcd DB exceeds this
	LeaderChangesLimit int  `json:"leaderChangesLimit,omitempty"` // alert if leader changes > this per run
}

type APIServerCheckSpec struct {
	Enabled            bool `json:"enabled"`
	LatencyThresholdMs int  `json:"latencyThresholdMs,omitempty"`
	ErrorRatePercent   int  `json:"errorRatePercent,omitempty"`
}

type NodeCheckSpec struct {
	Enabled             bool `json:"enabled"`
	CheckDiskPressure   bool `json:"checkDiskPressure,omitempty"`
	CheckMemoryPressure bool `json:"checkMemoryPressure,omitempty"`
	CheckPIDPressure    bool `json:"checkPIDPressure,omitempty"`
	CheckNotReady       bool `json:"checkNotReady,omitempty"`
}

type ClusterOperatorCheckSpec struct {
	Enabled bool `json:"enabled"`
}

type MCPCheckSpec struct {
	Enabled bool `json:"enabled"`
}

// PodSelectorSpec targets specific pods within a namespace
type PodSelectorSpec struct {
	Namespace     string `json:"namespace"`
	LabelSelector string `json:"labelSelector,omitempty"`
}

type InfraPodCheckSpec struct {
	Enabled      bool              `json:"enabled"`
	Namespaces   []string          `json:"namespaces,omitempty"`
	PodSelectors []PodSelectorSpec `json:"podSelectors,omitempty"`
}

type ArgoCDCheckSpec struct {
	Enabled              bool     `json:"enabled"`
	Namespace            string   `json:"namespace,omitempty"` // default: openshift-gitops
	CheckApplications    bool     `json:"checkApplications,omitempty"`
	CheckApplicationSets bool     `json:"checkApplicationSets,omitempty"`
	AlertOn              []string `json:"alertOn,omitempty"` // Degraded, Unknown, OutOfSync
}

// CustomProjectSpec defines a user-defined namespace to monitor
type CustomProjectSpec struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	CheckDeployments  bool   `json:"checkDeployments,omitempty"`
	CheckStatefulSets bool   `json:"checkStatefulSets,omitempty"`
	CheckDaemonSets   bool   `json:"checkDaemonSets,omitempty"`
}

type CustomProjectCheckSpec struct {
	Enabled  bool                `json:"enabled"`
	Projects []CustomProjectSpec `json:"projects,omitempty"`
}

// ─── Notification specs ─────────────────────────────────────────────────────

type EmailNotificationSpec struct {
	Enabled  bool     `json:"enabled"`
	SMTP     string   `json:"smtp,omitempty"`
	From     string   `json:"from,omitempty"`
	To       []string `json:"to,omitempty"`
	OnStatus []string `json:"onStatus,omitempty"` // Always / Warning / Critical
}

type SecretKeyRef struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type SlackNotificationSpec struct {
	Enabled          bool         `json:"enabled"`
	WebhookURL      string       `json:"webhookURL,omitempty"`
	WebhookSecretRef SecretKeyRef `json:"webhookSecretRef,omitempty"`
	OnStatus         []string     `json:"onStatus,omitempty"`
}

type WebhookNotificationSpec struct {
	Enabled  bool     `json:"enabled"`
	URL      string   `json:"url,omitempty"`
	OnStatus []string `json:"onStatus,omitempty"`
}

type NotificationSpec struct {
	Email   EmailNotificationSpec   `json:"email,omitempty"`
	Slack   SlackNotificationSpec   `json:"slack,omitempty"`
	Webhook WebhookNotificationSpec `json:"webhook,omitempty"`
}

// ─── Checks aggregator ──────────────────────────────────────────────────────

type ChecksSpec struct {
	Etcd             EtcdCheckSpec            `json:"etcd,omitempty"`
	APIServer        APIServerCheckSpec       `json:"apiServer,omitempty"`
	Nodes            NodeCheckSpec            `json:"nodes,omitempty"`
	ClusterOperators ClusterOperatorCheckSpec `json:"clusterOperators,omitempty"`
	MCPs             MCPCheckSpec             `json:"mcps,omitempty"`
	InfraPods        InfraPodCheckSpec        `json:"infraPods,omitempty"`
	ArgoCD           ArgoCDCheckSpec          `json:"argocd,omitempty"`
	CustomProjects   CustomProjectCheckSpec   `json:"customProjects,omitempty"`
}

// ─── Reporting spec ─────────────────────────────────────────────────────────

type ReportingSpec struct {
	PrometheusMetrics bool `json:"prometheusMetrics,omitempty"`
	StoreReports      bool `json:"storeReports,omitempty"`
	RetentionDays     int  `json:"retentionDays,omitempty"`
}

// ─── Main Spec / Status ─────────────────────────────────────────────────────

type ClusterHealthCheckSpec struct {
	Schedule      string           `json:"schedule"` // cron expression
	Checks        ChecksSpec       `json:"checks,omitempty"`
	Notifications NotificationSpec `json:"notifications,omitempty"`
	Reporting     ReportingSpec    `json:"reporting,omitempty"`
}

type ClusterHealthCheckStatus struct {
	LastRunTime    *metav1.Time `json:"lastRunTime,omitempty"`
	LastRunStatus  string       `json:"lastRunStatus,omitempty"` // OK / Warning / Critical
	LastReportName string       `json:"lastReportName,omitempty"`
	Message        string       `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
// +kubebuilder:printcolumn:name="LastStatus",type=string,JSONPath=`.status.lastRunStatus`
// +kubebuilder:printcolumn:name="LastRun",type=string,JSONPath=`.status.lastRunTime`

type ClusterHealthCheck struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterHealthCheckSpec   `json:"spec,omitempty"`
	Status ClusterHealthCheckStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ClusterHealthCheckList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterHealthCheck `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterHealthCheck{}, &ClusterHealthCheckList{})
}
