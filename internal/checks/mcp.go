package checks

import (
	"context"
	"fmt"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RunMCPCheck(
	ctx context.Context,
	c client.Client,
) *healthcheckv1alpha1.MCPCheckResult {

	result := &healthcheckv1alpha1.MCPCheckResult{Status: "OK"}

	mcpList := &unstructured.UnstructuredList{}
	mcpList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "machineconfiguration.openshift.io",
		Version: "v1",
		Kind:    "MachineConfigPoolList",
	})

	if err := c.List(ctx, mcpList); err != nil {
		result.Status = "OK"
		result.Message = "MCPs not available (non-OpenShift cluster)"
		return result
	}

	for _, mcp := range mcpList.Items {
		name := mcp.GetName()

		machineCount, _, _ := unstructured.NestedInt64(mcp.Object, "status", "machineCount")
		readyCount, _, _ := unstructured.NestedInt64(mcp.Object, "status", "readyMachineCount")
		degradedCount, _, _ := unstructured.NestedInt64(mcp.Object, "status", "degradedMachineCount")

		detail := healthcheckv1alpha1.MCPDetail{
			Name:                 name,
			MachineCount:         int(machineCount),
			ReadyMachineCount:    int(readyCount),
			DegradedMachineCount: int(degradedCount),
			Status:               "OK",
		}

		if degradedCount > 0 {
			detail.Status = "Warning"
			result.Status = "Warning"
		}

		result.Pools = append(result.Pools, detail)
	}

	if result.Status == "OK" {
		result.Message = fmt.Sprintf("All %d MachineConfigPool(s) are healthy", len(result.Pools))
	} else {
		result.Message = "One or more MCPs have degraded machines"
	}

	return result
}
