# cluster-health-operator

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)
![Kubernetes](https://img.shields.io/badge/kubernetes-1.25+-blue.svg)
![OpenShift](https://img.shields.io/badge/openshift-4.10+-red.svg)
![Status](https://img.shields.io/badge/status-alpha-orange.svg)
<<<<<<< HEAD
=======

A production-grade Kubernetes/OpenShift Operator for unified cluster health monitoring.

## Problem

No existing operator covers all major cluster health domains in one 
configurable, schedulable, CRD-driven package:

| Existing Tool | Gap |
|---|---|
| Insights Operator | Cloud-connected only, no custom scope |
| Cluster Observability Operator | Control plane only, no ArgoCD/MCPs |
| K8sGPT | AI narrative, not structured scheduled reports |
| Node Health Check | Nodes + remediation only |

`cluster-health-operator` was built to fill this gap.

## Overview

`cluster-health-operator` fills a gap in the existing ecosystem — no single operator covers all major cluster components in one configurable, schedulable, CRD-driven package.

It runs scheduled health checks across all critical cluster domains and produces structured `ClusterHealthReport` resources, with opt-in notifications via Email, Slack, or Webhook.

## What It Checks

| Domain | Details |
|---|---|
| **etcd** | Member count, quorum, leader, pod health |
| **API Server** | Pod health, crash detection |
| **Nodes** | Ready state, DiskPressure, MemoryPressure, PIDPressure |
| **ClusterOperators** | Degraded state detection (OpenShift) |
| **MachineConfigPools** | Degraded machine count (OpenShift) |
| **Infra Pods** | Configurable namespace scanning, CrashLoopBackOff detection |
| **ArgoCD Applications** | Health status, sync status, configurable alert conditions |
| **Custom Projects** | User-defined namespaces with deployment/statefulset/daemonset checks |

## Architecture
ClusterHealthCheck CR (config)
│
▼
Controller (cron-scheduled)
│
├── etcd checker
├── APIServer checker
├── Node checker
├── ClusterOperator checker
├── MCP checker
├── InfraPod checker
├── ArgoCD checker
└── CustomProject checker
│
▼
ClusterHealthReport CR (results)
│
▼
Notifier (Email / Slack / Webhook)

## Installation

### Prerequisites
- Kubernetes 1.25+ or OpenShift 4.10+
- kubectl / oc
- operator-sdk (for development)

### Deploy

```bash
# Install CRDs
make install

# Deploy operator
make deploy IMG=cluster-health-operator:v0.1.0
```

### Quick Start

```yaml
apiVersion: healthcheck.io/v1alpha1
kind: ClusterHealthCheck
metadata:
  name: daily-health
spec:
  schedule: "0 7 * * *"    # every day at 7 AM

  checks:
    etcd:
      enabled: true
    apiServer:
      enabled: true
    nodes:
      enabled: true
      checkDiskPressure: true
      checkMemoryPressure: true
    clusterOperators:
      enabled: true
    mcps:
      enabled: true
    infraPods:
      enabled: true
      namespaces:
        - openshift-etcd
        - openshift-kube-apiserver
        - openshift-ingress
        - openshift-monitoring
    argocd:
      enabled: true
      namespace: openshift-gitops
      alertOn:
        - Degraded
        - Unknown

  notifications:
    slack:
      enabled: true
      webhookURL: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
      onStatus:
        - Warning
        - Critical
    email:
      enabled: false

  reporting:
    storeReports: true
    retentionDays: 30
```

```bash
kubectl apply -f your-healthcheck.yaml
```

### View Results

```bash
# Check status
kubectl get clusterhealthchecks

# List reports
kubectl get clusterhealthreports

# See full report
kubectl describe clusterhealthreport <report-name>
```

## Notifications

All notifiers are **opt-in** and controlled via `spec.notifications` in the CR.

| Notifier | Config field | onStatus values |
|---|---|---|
| Email | `notifications.email` | Always / Warning / Critical |
| Slack | `notifications.slack` | Always / Warning / Critical |
| Webhook | `notifications.webhook` | Always / Warning / Critical |

If `onStatus` is empty, notifications are always sent.

## Report Status Values

| Status | Meaning |
|---|---|
| `OK` | All checks passed |
| `Warning` | One or more checks degraded but non-critical |
| `Critical` | Quorum loss or complete component failure |

## Compatibility

| Platform | ClusterOperators | MCPs | ArgoCD | etcd | Nodes |
|---|---|---|---|---|---|
| OpenShift 4.x | ✅ | ✅ | ✅ | ✅ | ✅ |
| Vanilla Kubernetes | ⏭ skipped | ⏭ skipped | ✅ | ✅ | ✅ |

## Development

```bash
# Run locally against current kubeconfig context
make run

# Build image
docker build -t cluster-health-operator:v0.1.0 .

# Run tests
make test
```

## Roadmap

- [ ] Prometheus metrics exposure
- [ ] HyperShift hosted cluster awareness
- [ ] ACM multi-cluster support
- [ ] Web UI dashboard
- [ ] OLM bundle for OperatorHub

## License

Apache 2.0
>>>>>>> 87bb5b8a6251c877850110bce77a316991e71281
