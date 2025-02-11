package services

import (
	"context"
	"fmt"
	"log"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"os"
	"path/filepath"
	"strings"
)

type EmailPayload struct {
	APIKey    string `json:"apikey"`
	Namespace string `json:"namespace"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

func SendNotifications(threshold models.Threshold, dataName string, percentChange float64, recipients []models.Recipient, userEmail string) {
	log.Println("📨 Preparing mock notifications for threshold ID:", threshold.ThresholdID)

	if len(recipients) > 0 {
		thresholdExceeded := percentChange > threshold.ThresholdValue
		var emailTemplate string
		if thresholdExceeded {
			emailTemplate = "recipient_notification_bad.txt"
		} else {
			emailTemplate = "recipient_notification_good.txt"
		}

		for _, recipient := range recipients {
			subject := fmt.Sprintf("Urgent: %s Economic Data Alert", dataName)
			message, err := formatEmailFromTemplate(emailTemplate, map[string]string{
				"Recipient Name":    recipient.FirstName + " " + recipient.LastName,
				"Threshold Name":    dataName,
				"Change Percentage": fmt.Sprintf("%.2f", percentChange),
				"User First Name":   os.Getenv("SENDER_FIRST_NAME"),
				"User Last Name":    os.Getenv("SENDER_LAST_NAME"),
				"User Email":        os.Getenv("SENDER_EMAIL"),
			})
			if err != nil {
				log.Printf("❌ Error formatting recipient email: %v", err)
				continue
			}

			log.Printf("📧 [MOCK EMAIL] To: %s | Subject: %s", recipient.Email, subject)
			log.Println("📧 Email Body:")
			log.Println(message)
		}
	}

	if threshold.NotifyUser {
		userMessage, err := formatEmailFromTemplate("user_notification.txt", map[string]string{
			"User First Name":   os.Getenv("SENDER_FIRST_NAME"),
			"Threshold Name":    dataName,
			"New Value":         fmt.Sprintf("%.2f", percentChange),
			"Threshold Value":   fmt.Sprintf("%.2f", threshold.ThresholdValue),
			"Change Percentage": fmt.Sprintf("%.2f", percentChange),
			"Good/Bad":          determineChangeDirection(percentChange, threshold.ThresholdValue),
			"Recipient List":    formatRecipientList(recipients),
		})
		if err != nil {
			log.Printf("❌ Error formatting user email: %v", err)
			return
		}

		log.Printf("📧 [MOCK EMAIL] To: %s | Subject: %s", userEmail, "Your MEGGA Threshold Was Hit - Here's What to Do Next")
		log.Println("📧 Email Body:")
		log.Println(userMessage)
	}
}

func formatRecipientList(recipients []models.Recipient) string {
	var recipientList strings.Builder
	for _, r := range recipients {
		recipientList.WriteString(fmt.Sprintf("%s %s <%s>\n", r.FirstName, r.LastName, r.Email))
	}
	return recipientList.String()
}

func determineChangeDirection(percentChange, thresholdValue float64) (direction string) {
	if percentChange > thresholdValue {
		direction = "bad"
	} else {
		direction = "good"
	}
	return
}

func formatEmailFromTemplate(templateFile string, replacements map[string]string) (string, error) {
	templatePath := filepath.Clean(filepath.Join("internal", "services", "email_templates", templateFile))

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read email template %s: %w", templateFile, err)
	}

	message := string(content)
	for placeholder, value := range replacements {
		message = strings.ReplaceAll(message, "["+placeholder+"]", value)
	}
	return message, nil
}

func fetchRecipientsForThreshold(db database.DBQuerier, thresholdID int) ([]models.Recipient, error) {
	var recipients []models.Recipient
	rows, err := db.Query(context.Background(),
		`SELECT r.recipient_id, r.email, r.first_name, r.last_name, r.designation
		FROM recipients r
		JOIN threshold_recipients tr ON r.recipient_id = tr.recipient_id
		WHERE tr.threshold_id = $1`, thresholdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var recipient models.Recipient
		if err := rows.Scan(&recipient.RecipientID, &recipient.Email, &recipient.FirstName, &recipient.LastName, &recipient.Designation); err != nil {
			return nil, err
		}
		recipients = append(recipients, recipient)
	}
	return recipients, nil
}

func fetchUserEmail(db database.DBQuerier, userID int) string {
	var email string
	err := db.QueryRow(context.Background(), "SELECT email FROM users WHERE user_id = $1", userID).Scan(&email)
	if err != nil {
		log.Printf("❌ Failed to fetch user email for user ID %d: %v", userID, err)
		return ""
	}
	return email
}
