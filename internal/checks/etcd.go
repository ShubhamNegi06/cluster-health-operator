package checks

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
)

func RunEtcdCheck(
	ctx context.Context,
	c client.Client,
	spec healthcheckv1alpha1.EtcdCheckSpec,
) *healthcheckv1alpha1.EtcdCheckResult {

	result := &healthcheckv1alpha1.EtcdCheckResult{Status: "OK"}

	// Try OpenShift namespace first, fall back to kube-system
	namespacesToTry := []string{"openshift-etcd", "kube-system"}

	var etcdPods []corev1.Pod

	for _, ns := range namespacesToTry {
		podList := &corev1.PodList{}
		if err := c.List(ctx, podList,
			client.InNamespace(ns),
			client.MatchingLabels{"component": "etcd"},
		); err != nil || len(podList.Items) == 0 {
			// try alternate label used by kubeadm/kind
			podList2 := &corev1.PodList{}
			_ = c.List(ctx, podList2,
				client.InNamespace(ns),
				client.MatchingLabels{"tier": "control-plane", "component": "etcd"},
			)
			if len(podList2.Items) > 0 {
				etcdPods = append(etcdPods, podList2.Items...)
				break
			}
			continue
		}
		etcdPods = append(etcdPods, podList.Items...)
		break
	}

	if len(etcdPods) == 0 {
		result.Status = "Warning"
		result.Message = "No etcd pods found — check namespace or labels"
		return result
	}

	result.MemberCount = len(etcdPods)
	healthyCount := 0
	var leader string

	for _, pod := range etcdPods {
		podHealthy := true

		// Check pod phase
		if pod.Status.Phase != corev1.PodRunning {
			podHealthy = false
			result.Status = "Warning"
		}

		// Check container statuses
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				podHealthy = false
				result.Status = "Warning"
			}
			if cs.RestartCount > 5 {
				result.Status = "Warning"
			}
		}

		if podHealthy {
			healthyCount++
			if leader == "" {
				leader = pod.Name // first healthy pod noted as leader candidate
			}
		}
	}

	result.HealthyMembers = healthyCount
	result.Leader = leader

	// Quorum check — majority must be healthy
	quorum := (result.MemberCount / 2) + 1
	if healthyCount < quorum {
		result.Status = "Critical"
		result.Message = fmt.Sprintf(
			"etcd quorum at risk — only %d of %d members healthy (need %d)",
			healthyCount, result.MemberCount, quorum,
		)
		return result
	}

	if result.Status == "OK" {
		result.Message = fmt.Sprintf(
			"etcd healthy — %d/%d members running",
			healthyCount, result.MemberCount,
		)
	} else {
		result.Message = fmt.Sprintf(
			"etcd degraded — %d/%d members healthy",
			healthyCount, result.MemberCount,
		)
	}

	return result
}
