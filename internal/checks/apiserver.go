package checks

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
)

func RunAPIServerCheck(
	ctx context.Context,
	c client.Client,
	spec healthcheckv1alpha1.APIServerCheckSpec,
) *healthcheckv1alpha1.APIServerCheckResult {

	result := &healthcheckv1alpha1.APIServerCheckResult{Status: "OK"}

	// Try OpenShift namespace first, fall back to kube-system
	namespacesToTry := []string{"openshift-kube-apiserver", "kube-system"}

	var apiserverPods []corev1.Pod

	for _, ns := range namespacesToTry {
		podList := &corev1.PodList{}
		// OpenShift label
		if err := c.List(ctx, podList,
			client.InNamespace(ns),
			client.MatchingLabels{"app": "openshift-kube-apiserver"},
		); err == nil && len(podList.Items) > 0 {
			apiserverPods = append(apiserverPods, podList.Items...)
			break
		}

		// kubeadm/kind label
		podList2 := &corev1.PodList{}
		if err := c.List(ctx, podList2,
			client.InNamespace(ns),
			client.MatchingLabels{"component": "kube-apiserver"},
		); err == nil && len(podList2.Items) > 0 {
			apiserverPods = append(apiserverPods, podList2.Items...)
			break
		}
	}

	if len(apiserverPods) == 0 {
		// API server is obviously running if we got this far —
		// we just couldn't find its pod. Report OK with a note.
		result.Status = "OK"
		result.Message = "API server is responsive (pod not found by label — may be managed differently)"
		return result
	}

	totalPods := len(apiserverPods)
	healthyPods := 0

	for _, pod := range apiserverPods {
		podHealthy := true

		if pod.Status.Phase != corev1.PodRunning {
			podHealthy = false
			result.Status = "Warning"
		}

		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				podHealthy = false
				result.Status = "Warning"
			}
			if cs.State.Waiting != nil {
				reason := cs.State.Waiting.Reason
				if reason == "CrashLoopBackOff" || reason == "Error" {
					result.Status = "Critical"
				}
			}
		}

		if podHealthy {
			healthyPods++
		}
	}

	if result.Status == "OK" {
		result.Message = fmt.Sprintf(
			"API server healthy — %d/%d pods running",
			healthyPods, totalPods,
		)
	} else {
		result.Message = fmt.Sprintf(
			"API server degraded — %d/%d pods healthy",
			healthyPods, totalPods,
		)
	}

	return result
}
