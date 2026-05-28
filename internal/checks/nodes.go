package checks

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
)

func RunNodeCheck(
	ctx context.Context,
	c client.Client,
	spec healthcheckv1alpha1.NodeCheckSpec,
) *healthcheckv1alpha1.NodeCheckResult {

	result := &healthcheckv1alpha1.NodeCheckResult{Status: "OK"}

	nodeList := &corev1.NodeList{}
	if err := c.List(ctx, nodeList); err != nil {
		result.Status = "Critical"
		result.Message = fmt.Sprintf("Failed to list nodes: %v", err)
		return result
	}

	result.Total = len(nodeList.Items)

	for _, node := range nodeList.Items {
		ready := false

		for _, cond := range node.Status.Conditions {
			// ── Not Ready ───────────────────────────────────────────────
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					ready = true
				} else if spec.CheckNotReady {
					result.NotReady = append(result.NotReady, healthcheckv1alpha1.NodeDetail{
						Name:   node.Name,
						Reason: cond.Reason,
					})
				}
			}

			// ── Disk Pressure ────────────────────────────────────────────
			if spec.CheckDiskPressure &&
				cond.Type == corev1.NodeDiskPressure &&
				cond.Status == corev1.ConditionTrue {
				result.NotReady = append(result.NotReady, healthcheckv1alpha1.NodeDetail{
					Name:   node.Name,
					Reason: "DiskPressure",
				})
			}

			// ── Memory Pressure ──────────────────────────────────────────
			if spec.CheckMemoryPressure &&
				cond.Type == corev1.NodeMemoryPressure &&
				cond.Status == corev1.ConditionTrue {
				result.NotReady = append(result.NotReady, healthcheckv1alpha1.NodeDetail{
					Name:   node.Name,
					Reason: "MemoryPressure",
				})
			}

			// ── PID Pressure ─────────────────────────────────────────────
			if spec.CheckPIDPressure &&
				cond.Type == corev1.NodePIDPressure &&
				cond.Status == corev1.ConditionTrue {
				result.NotReady = append(result.NotReady, healthcheckv1alpha1.NodeDetail{
					Name:   node.Name,
					Reason: "PIDPressure",
				})
			}
		}

		if ready {
			result.Ready++
		}
	}

	// ── Determine Status ─────────────────────────────────────────────────
	if len(result.NotReady) > 0 {
		result.Status = "Warning"
		result.Message = fmt.Sprintf("%d node(s) have issues", len(result.NotReady))
	} else {
		result.Message = fmt.Sprintf("All %d nodes are healthy", result.Total)
	}

	return result
}
