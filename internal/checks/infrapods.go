package checks

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
)

func RunInfraPodCheck(
	ctx context.Context,
	c client.Client,
	spec healthcheckv1alpha1.InfraPodCheckSpec,
) *healthcheckv1alpha1.InfraPodCheckResult {

	result := &healthcheckv1alpha1.InfraPodCheckResult{Status: "OK"}

	// collect all namespaces to check
	namespaces := spec.Namespaces

	// add namespaces from podSelectors too
	for _, ps := range spec.PodSelectors {
		namespaces = append(namespaces, ps.Namespace)
	}

	// deduplicate
	seen := map[string]bool{}
	unique := []string{}
	for _, ns := range namespaces {
		if !seen[ns] {
			seen[ns] = true
			unique = append(unique, ns)
		}
	}

	result.CheckedNamespaces = len(unique)

	for _, ns := range unique {
		podList := &corev1.PodList{}
		if err := c.List(ctx, podList, client.InNamespace(ns)); err != nil {
			continue
		}

		result.TotalPodsChecked += len(podList.Items)

		for _, pod := range podList.Items {
			if !isPodHealthy(pod) {
				result.UnhealthyPods = append(result.UnhealthyPods,
					healthcheckv1alpha1.UnhealthyPod{
						Namespace: pod.Namespace,
						Name:      pod.Name,
						Phase:     string(pod.Status.Phase),
						Reason:    pod.Status.Reason,
					})
			}
		}
	}

	if len(result.UnhealthyPods) > 0 {
		result.Status = "Warning"
		result.Message = fmt.Sprintf("%d unhealthy pod(s) found across %d namespace(s)",
			len(result.UnhealthyPods), result.CheckedNamespaces)
	} else {
		result.Message = fmt.Sprintf("All %d pods healthy across %d namespace(s)",
			result.TotalPodsChecked, result.CheckedNamespaces)
	}

	return result
}

func isPodHealthy(pod corev1.Pod) bool {
	// Completed jobs are fine
	if pod.Status.Phase == corev1.PodSucceeded {
		return true
	}
	// Running or pending is OK
	if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
		// check for crash looping containers
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.RestartCount > 5 {
				return false
			}
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "CrashLoopBackOff" || reason == "OOMKilled" || reason == "Error" {
					return false
				}
			}
		}
		return true
	}
	return false
}
