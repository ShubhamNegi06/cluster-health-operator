package checks

import (
	"context"
	"time"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func CleanupOldReports(
	ctx context.Context,
	c client.Client,
	triggeredBy string,
	retentionDays int,
) {
	logger := log.FromContext(ctx)

	if retentionDays <= 0 {
		retentionDays = 30 // default
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)

	reportList := &healthcheckv1alpha1.ClusterHealthReportList{}
	if err := c.List(ctx, reportList); err != nil {
		logger.Error(err, "Failed to list ClusterHealthReports for cleanup")
		return
	}

	deleted := 0
	for _, report := range reportList.Items {
		// Only clean up reports from this ClusterHealthCheck
		if report.Spec.TriggeredBy != triggeredBy {
			continue
		}
		if report.Spec.RunTime != nil && report.Spec.RunTime.Time.Before(cutoff) {
			if err := c.Delete(ctx, &report); err != nil {
				logger.Error(err, "Failed to delete old report", "name", report.Name)
			} else {
				deleted++
				logger.Info("Deleted old report", "name", report.Name, "age", report.Spec.RunTime.Time)
			}
		}
	}

	if deleted > 0 {
		logger.Info("Cleaned up old reports", "deleted", deleted, "triggeredBy", triggeredBy)
	}
}
