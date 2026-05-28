package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type webhookPayload struct {
	ReportName    string `json:"reportName"`
	OverallStatus string `json:"overallStatus"`
	Summary       string `json:"summary"`
	Timestamp     string `json:"timestamp"`
	Source        string `json:"source"`
}

func SendWebhook(
	url string,
	reportName string,
	overallStatus string,
	summary string,
) error {
	payload := webhookPayload{
		ReportName:    reportName,
		OverallStatus: overallStatus,
		Summary:       summary,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Source:        "cluster-health-operator",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	return nil
}
