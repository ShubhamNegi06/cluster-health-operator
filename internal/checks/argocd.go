package checks

import (
	"context"
	"fmt"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var argoAppGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

func RunArgoCDCheck(
	ctx context.Context,
	c client.Client,
	spec healthcheckv1alpha1.ArgoCDCheckSpec,
) *healthcheckv1alpha1.ArgoCDCheckResult {

	result := &healthcheckv1alpha1.ArgoCDCheckResult{Status: "OK"}

	ns := spec.Namespace
	if ns == "" {
		ns = "openshift-gitops"
	}

	appList := &unstructured.UnstructuredList{}
	appList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "argoproj.io",
		Version: "v1alpha1",
		Kind:    "ApplicationList",
	})

	if err := c.List(ctx, appList, client.InNamespace(ns)); err != nil {
		result.Status = "OK"
		result.Message = fmt.Sprintf("ArgoCD not available in namespace %s", ns)
		return result
	}

	result.TotalApps = len(appList.Items)

	alertOn := map[string]bool{}
	for _, s := range spec.AlertOn {
		alertOn[s] = true
	}

	for _, app := range appList.Items {
		name := app.GetName()

		healthStatus, _, _ := unstructured.NestedString(app.Object, "status", "health", "status")
		syncStatus, _, _ := unstructured.NestedString(app.Object, "status", "sync", "status")
		message, _, _ := unstructured.NestedString(app.Object, "status", "conditions")

		appDetail := healthcheckv1alpha1.ArgoCDAppDetail{
			Name:         name,
			HealthStatus: healthStatus,
			SyncStatus:   syncStatus,
			Message:      message,
		}

		if alertOn["Degraded"] && healthStatus == "Degraded" {
			result.Degraded = append(result.Degraded, appDetail)
			result.Status = "Warning"
		} else if alertOn["Unknown"] && healthStatus == "Unknown" {
			result.Degraded = append(result.Degraded, appDetail)
			result.Status = "Warning"
		} else if alertOn["OutOfSync"] && syncStatus == "OutOfSync" {
			result.OutOfSync = append(result.OutOfSync, appDetail)
			result.Status = "Warning"
		} else {
			result.Healthy++
		}
	}

	if result.Status == "OK" {
		result.Message = fmt.Sprintf("All %d ArgoCD application(s) are healthy", result.TotalApps)
	} else {
		result.Message = fmt.Sprintf("%d degraded, %d out-of-sync ArgoCD application(s)",
			len(result.Degraded), len(result.OutOfSync))
	}

	return result
}
