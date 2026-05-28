package notifier

import (
	"context"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// SendNotifications dispatches to all enabled notifiers based on CR spec
func SendNotifications(
	ctx context.Context,
	spec healthcheckv1alpha1.NotificationSpec,
	reportName string,
	overallStatus string,
	summary string,
) {
	logger := log.FromContext(ctx)

	// ── Email ────────────────────────────────────────────────────────────
	if spec.Email.Enabled {
		if shouldNotify(spec.Email.OnStatus, overallStatus) {
			logger.Info("Sending email notification", "status", overallStatus)
			if err := SendEmail(spec.Email, reportName, overallStatus, summary); err != nil {
				logger.Error(err, "Email notification failed")
			} else {
				logger.Info("Email notification sent")
			}
		} else {
			logger.Info("Email skipped — status does not match OnStatus filter",
				"status", overallStatus, "onStatus", spec.Email.OnStatus)
		}
	}

	// ── Slack ────────────────────────────────────────────────────────────
	if spec.Slack.Enabled {
		if shouldNotify(spec.Slack.OnStatus, overallStatus) {
			logger.Info("Sending Slack notification", "status", overallStatus)
			webhookURL := spec.Slack.WebhookURL
			if webhookURL == "" {
				logger.Error(nil, "Slack enabled but webhookURL is empty")
			} else {
				if err := SendSlack(webhookURL, reportName, overallStatus, summary); err != nil {
					logger.Error(err, "Slack notification failed")
				} else {
					logger.Info("Slack notification sent")
				}
			}
		} else {
			logger.Info("Slack skipped — status does not match OnStatus filter",
				"status", overallStatus, "onStatus", spec.Slack.OnStatus)
		}
	}

	// ── Webhook ──────────────────────────────────────────────────────────
	if spec.Webhook.Enabled {
		if shouldNotify(spec.Webhook.OnStatus, overallStatus) {
			logger.Info("Sending Webhook notification", "status", overallStatus)
			if spec.Webhook.URL == "" {
				logger.Error(nil, "Webhook enabled but URL is empty")
			} else {
				if err := SendWebhook(spec.Webhook.URL, reportName, overallStatus, summary); err != nil {
					logger.Error(err, "Webhook notification failed")
				} else {
					logger.Info("Webhook notification sent")
				}
			}
		} else {
			logger.Info("Webhook skipped — status does not match OnStatus filter",
				"status", overallStatus, "onStatus", spec.Webhook.OnStatus)
		}
	}
}

// shouldNotify checks if the current status matches the user-defined filter
// If OnStatus is empty — always notify
// Supported values: "Always", "Warning", "Critical"
func shouldNotify(onStatus []string, currentStatus string) bool {
	if len(onStatus) == 0 {
		return true
	}
	for _, s := range onStatus {
		if s == "Always" {
			return true
		}
		if s == currentStatus {
			return true
		}
		// Critical also triggers Warning rules
		if s == "Warning" && currentStatus == "Critical" {
			return true
		}
	}
	return false
}
