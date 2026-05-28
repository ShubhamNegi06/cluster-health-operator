package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type slackPayload struct {
	Text        string            `json:"text"`
	Attachments []slackAttachment `json:"attachments"`
}

type slackAttachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
	Ts     int64  `json:"ts"`
}

func SendSlack(
	webhookURL string,
	reportName string,
	overallStatus string,
	summary string,
) error {
	color := colorForStatus(overallStatus)

	payload := slackPayload{
		Text: fmt.Sprintf("*Cluster Health Check — %s*", overallStatus),
		Attachments: []slackAttachment{
			{
				Color:  color,
				Title:  reportName,
				Text:   summary,
				Footer: "cluster-health-operator",
				Ts:     time.Now().Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func colorForStatus(status string) string {
	switch status {
	case "Critical":
		return "#FF0000" // red
	case "Warning":
		return "#FFA500" // orange
	default:
		return "#36A64F" // green
	}
}
