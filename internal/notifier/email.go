package notifier

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	healthcheckv1alpha1 "github.com/ShubhamNegi06/cluster-health-operator/api/v1alpha1"
)

func SendEmail(
	spec healthcheckv1alpha1.EmailNotificationSpec,
	reportName string,
	overallStatus string,
	summary string,
) error {
	if spec.SMTP == "" {
		return fmt.Errorf("SMTP address is not configured")
	}
	if len(spec.To) == 0 {
		return fmt.Errorf("no recipients configured")
	}

	subject := fmt.Sprintf("[ClusterHealth] %s — %s", overallStatus, reportName)

	body := fmt.Sprintf(`Cluster Health Check Report
===========================
Report   : %s
Status   : %s
Time     : %s

Summary
-------
%s

---
Sent by cluster-health-operator
`,
		reportName,
		overallStatus,
		time.Now().Format(time.RFC1123),
		summary,
	)

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		spec.From,
		strings.Join(spec.To, ", "),
		subject,
		body,
	)

	// Parse host:port
	parts := strings.Split(spec.SMTP, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid SMTP format, expected host:port, got: %s", spec.SMTP)
	}
	host := parts[0]

	// Try STARTTLS first, fall back to plain
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", spec.SMTP, tlsConfig)
	if err != nil {
		// Fall back to plain SMTP (port 25 or internal relay)
		return smtp.SendMail(
			spec.SMTP,
			nil, // no auth — for internal relay
			spec.From,
			spec.To,
			[]byte(message),
		)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	if err = client.Mail(spec.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, to := range spec.To {
		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("RCPT TO failed for %s: %w", to, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	return w.Close()
}
