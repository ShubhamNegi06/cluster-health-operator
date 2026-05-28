package checks

import (
	"context"
	"fmt"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RunClusterOperatorCheck(
	ctx context.Context,
	c client.Client,
) *healthcheckv1alpha1.ClusterOperatorCheckResult {

	result := &healthcheckv1alpha1.ClusterOperatorCheckResult{Status: "OK"}

	coList := &unstructured.UnstructuredList{}
	coList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.openshift.io",
		Version: "v1",
		Kind:    "ClusterOperatorList",
	})

	if err := c.List(ctx, coList); err != nil {
		result.Status = "OK"
		result.Message = "ClusterOperators not available (non-OpenShift cluster)"
		return result
	}

	result.Total = len(coList.Items)

	for _, co := range coList.Items {
		name := co.GetName()
		conditions, found, err := unstructured.NestedSlice(co.Object, "status", "conditions")
		if err != nil || !found {
			continue
		}

		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			condType, _, _ := unstructured.NestedString(cond, "type")
			condStatus, _, _ := unstructured.NestedString(cond, "status")
			condMessage, _, _ := unstructured.NestedString(cond, "message")

			if condType == "Degraded" && condStatus == "True" {
				result.Degraded = append(result.Degraded, healthcheckv1alpha1.DegradedOperator{
					Name:    name,
					Message: condMessage,
				})
			}
		}
	}

	if len(result.Degraded) > 0 {
		result.Status = "Warning"
		result.Message = fmt.Sprintf("%d ClusterOperator(s) degraded", len(result.Degraded))
	} else {
		result.Message = fmt.Sprintf("All %d ClusterOperators are healthy", result.Total)
	}

	return result
}
