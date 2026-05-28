package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
	"github.com/ShubhamNegi06/cluster-health-operator/internal/checks"
	"github.com/ShubhamNegi06/cluster-health-operator/internal/notifier"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterHealthCheckReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	CronRunner    *cron.Cron
	ScheduledJobs map[string]cron.EntryID
	// tracks the schedule string per CR so we only reschedule on actual change
	ScheduledSpec map[string]string
}

func (r *ClusterHealthCheckReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the CR
	chc := &healthcheckv1alpha1.ClusterHealthCheck{}
	if err := r.Get(ctx, req.NamespacedName, chc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	currentSchedule := chc.Spec.Schedule
	previousSchedule, alreadyScheduled := r.ScheduledSpec[req.Name]

	// If schedule hasn't changed and already running — do nothing
	if alreadyScheduled && previousSchedule == currentSchedule {
		logger.Info("Schedule unchanged, skipping reschedule", "name", req.Name)
		return ctrl.Result{}, nil
	}

	// Remove old cron entry if exists
	if id, exists := r.ScheduledJobs[req.Name]; exists {
		r.CronRunner.Remove(id)
		delete(r.ScheduledJobs, req.Name)
		delete(r.ScheduledSpec, req.Name)
	}

	// Schedule new cron entry
	entryID, err := r.CronRunner.AddFunc(currentSchedule, func() {
		r.runHealthChecks(chc)
	})
	if err != nil {
		logger.Error(err, "Failed to schedule health check", "schedule", currentSchedule)
		return ctrl.Result{}, err
	}

	r.ScheduledJobs[req.Name] = entryID
	r.ScheduledSpec[req.Name] = currentSchedule
	logger.Info("Scheduled health check", "name", req.Name, "schedule", currentSchedule)

	// Run once immediately only on first-time creation (not already scheduled before)
	if !alreadyScheduled {
		logger.Info("First time scheduling — running initial check", "name", req.Name)
		go r.runHealthChecks(chc)
	}

	return ctrl.Result{}, nil
}

func (r *ClusterHealthCheckReconciler) runHealthChecks(chc *healthcheckv1alpha1.ClusterHealthCheck) {
	ctx := context.Background()
	logger := log.FromContext(ctx)
	logger.Info("Running health checks", "name", chc.Name)

	results := &healthcheckv1alpha1.CheckResults{}
	overallStatus := "OK"

	// ── Node Check ──────────────────────────────────────────────────────
	if chc.Spec.Checks.Nodes.Enabled {
		result := checks.RunNodeCheck(ctx, r.Client, chc.Spec.Checks.Nodes)
		results.Nodes = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("Node check done", "status", result.Status)
	}

	// ── Etcd Check ──────────────────────────────────────────────────────
	if chc.Spec.Checks.Etcd.Enabled {
		result := checks.RunEtcdCheck(ctx, r.Client, chc.Spec.Checks.Etcd)
		results.Etcd = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("Etcd check done", "status", result.Status)
	}

	// ── APIServer Check ──────────────────────────────────────────────────
	if chc.Spec.Checks.APIServer.Enabled {
		result := checks.RunAPIServerCheck(ctx, r.Client, chc.Spec.Checks.APIServer)
		results.APIServer = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("APIServer check done", "status", result.Status)
	}

	// ── ClusterOperator Check ────────────────────────────────────────────
	if chc.Spec.Checks.ClusterOperators.Enabled {
		result := checks.RunClusterOperatorCheck(ctx, r.Client)
		results.ClusterOperators = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("ClusterOperator check done", "status", result.Status)
	}

	// ── MCP Check ────────────────────────────────────────────────────────
	if chc.Spec.Checks.MCPs.Enabled {
		result := checks.RunMCPCheck(ctx, r.Client)
		results.MCPs = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("MCP check done", "status", result.Status)
	}

	// ── InfraPod Check ───────────────────────────────────────────────────
	if chc.Spec.Checks.InfraPods.Enabled {
		result := checks.RunInfraPodCheck(ctx, r.Client, chc.Spec.Checks.InfraPods)
		results.InfraPods = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("InfraPod check done", "status", result.Status)
	}

	// ── ArgoCD Check ─────────────────────────────────────────────────────
	if chc.Spec.Checks.ArgoCD.Enabled {
		result := checks.RunArgoCDCheck(ctx, r.Client, chc.Spec.Checks.ArgoCD)
		results.ArgoCD = result
		overallStatus = worstStatus(overallStatus, result.Status)
		logger.Info("ArgoCD check done", "status", result.Status)
	}

	// ── Write Report CR ──────────────────────────────────────────────────
	reportName := fmt.Sprintf("%s-%s", chc.Name, time.Now().UTC().Format("20060102-150405"))
	report := &healthcheckv1alpha1.ClusterHealthReport{
		ObjectMeta: metav1.ObjectMeta{
			Name: reportName,
		},
		Spec: healthcheckv1alpha1.ClusterHealthReportSpec{
			TriggeredBy: chc.Name,
			RunTime:     &metav1.Time{Time: time.Now()},
		},
	}

	if err := r.Create(ctx, report); err != nil {
		logger.Error(err, "Failed to create ClusterHealthReport")
		return
	}

	// Update report status
	report.Status.OverallStatus = overallStatus
	report.Status.Checks = *results
	report.Status.Summary = fmt.Sprintf("Health check completed at %s — Overall: %s",
		time.Now().Format(time.RFC3339), overallStatus)

	if err := r.Status().Update(ctx, report); err != nil {
		logger.Error(err, "Failed to update ClusterHealthReport status")
		return
	}

	// Re-fetch ClusterHealthCheck before updating status to avoid conflict
	fresh := &healthcheckv1alpha1.ClusterHealthCheck{}
	if err := r.Get(ctx, client.ObjectKey{Name: chc.Name}, fresh); err != nil {
		logger.Error(err, "Failed to re-fetch ClusterHealthCheck for status update")
		return
	}
	fresh.Status.LastRunTime = &metav1.Time{Time: time.Now()}
	fresh.Status.LastRunStatus = overallStatus
	fresh.Status.LastReportName = reportName
	if err := r.Status().Update(ctx, fresh); err != nil {
		logger.Error(err, "Failed to update ClusterHealthCheck status")
		return
	}

	notifier.SendNotifications(ctx, chc.Spec.Notifications, reportName, overallStatus, report.Status.Summary)

	checks.CleanupOldReports(ctx, r.Client, chc.Name, chc.Spec.Reporting.RetentionDays)

	logger.Info("Health check run complete", "overallStatus", overallStatus, "report", reportName)
}

func worstStatus(current, new string) string {
	rank := map[string]int{"OK": 0, "Warning": 1, "Critical": 2}
	if rank[new] > rank[current] {
		return new
	}
	return current
}

func (r *ClusterHealthCheckReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.CronRunner = cron.New()
	r.ScheduledJobs = make(map[string]cron.EntryID)
	r.ScheduledSpec = make(map[string]string)
	r.CronRunner.Start()

	return ctrl.NewControllerManagedBy(mgr).
		For(&healthcheckv1alpha1.ClusterHealthCheck{}).
		// CRITICAL: Only reconcile on Spec changes, not Status updates
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
