package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"megga-backend/internal/database"
	"megga-backend/internal/models"
	"net/http"
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

func SendThresholdNotifications(threshold models.Threshold, dataName string, percentChange float64) {
	log.Println("📨 Preparing notifications for threshold ID:", threshold.ThresholdID)

	thresholdExceeded := percentChange > threshold.ThresholdValue
	emailTemplate := "recipient_notification_negative.txt"
	if !thresholdExceeded {
		emailTemplate = "recipient_notification_positive.txt"
	}

	recipients, err := fetchRecipientsForThreshold(threshold.ThresholdID)
	if err != nil {
		log.Printf("❌ Failed to fetch recipients for threshold %d: %v", threshold.ThresholdID, err)
		return
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
		sendEmail(recipient.Email, subject, message)
	}

	if threshold.NotifyUser {
		userEmail := fetchUserEmail(threshold.UserID)
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
		sendEmail(userEmail, "Your MEGGA Threshold Was Hit - Here's What to Do Next", userMessage)
	}
}

// fetchRecipientsForThreshold retrieves all recipients tied to a threshold.
func fetchRecipientsForThreshold(thresholdID int) ([]models.Recipient, error) {
	var recipients []models.Recipient
	rows, err := database.DB.Query(context.Background(), `
		SELECT r.recipient_id, r.email, r.first_name, r.last_name, r.designation
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

// fetchUserEmail retrieves the user's email address based on user ID.
func fetchUserEmail(userID int) string {
	var email string
	err := database.DB.QueryRow(context.Background(), "SELECT email FROM users WHERE user_id = $1", userID).Scan(&email)
	if err != nil {
		log.Printf("❌ Failed to fetch user email for user ID %d: %v", userID, err)
		return ""
	}
	return email
}

// formatRecipientList formats the recipient list into a readable string.
func formatRecipientList(recipients []models.Recipient) string {
	var recipientList string
	for _, r := range recipients {
		recipientList += fmt.Sprintf("%s %s <%s>\n", r.FirstName, r.LastName, r.Email)
	}
	return recipientList
}

// determineChangeDirection determines whether the threshold change is good or bad.
func determineChangeDirection(percentChange, thresholdValue float64) string {
	if percentChange > thresholdValue {
		return "bad"
	}
	return "good"
}

// formatEmailFromTemplate reads an email template and replaces placeholders with actual values.
func formatEmailFromTemplate(templateFile string, replacements map[string]string) (string, error) {
	templatePath := filepath.Join("internal", "services", "email_templates", templateFile)
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read email template: %w", err)
	}

	message := string(content)
	for placeholder, value := range replacements {
		message = strings.ReplaceAll(message, "["+placeholder+"]", value)
	}
	return message, nil
}

// sendEmail sends an email using the Testmail.app API.
func sendEmail(to, subject, body string) error {
	apiKey := os.Getenv("TESTMAIL_API_KEY")
	namespace := os.Getenv("TESTMAIL_NAMESPACE")

	if apiKey == "" || namespace == "" {
		log.Println("❌ Missing Testmail.app API credentials. Check TESTMAIL_API_KEY and TESTMAIL_NAMESPACE.")
		return fmt.Errorf("missing Testmail.app API credentials")
	}

	emailPayload := EmailPayload{
		APIKey:    apiKey,
		Namespace: namespace,
		To:        to,
		Subject:   subject,
		Body:      body,
	}

	jsonData, err := json.Marshal(emailPayload)
	if err != nil {
		log.Printf("❌ Failed to marshal email payload: %v", err)
		return err
	}

	apiURL := "https://api.testmail.app/api/json/send"

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Email request returned non-200 status code: %d", resp.StatusCode)
		return fmt.Errorf("email request failed with status code: %d", resp.StatusCode)
	}

	log.Printf("📧 Successfully sent email to %s with subject: %s", to, subject)
	return nil
}


func SendNotifications(db database.DBQuerier, threshold models.Threshold, dataName string, percentChange float64) {
	log.Println("📨 Preparing notifications for threshold ID:", threshold.ThresholdID)

	recipients, err := fetchRecipientsForThreshold(threshold.ThresholdID)
	if err != nil {
		log.Printf("❌ Failed to fetch recipients for threshold %d: %v", threshold.ThresholdID, err)
		return
	}

	if len(recipients) > 0 {
		
	thresholdExceeded := percentChange > threshold.ThresholdValue
	emailTemplate := "recipient_notification_negative.txt"
	if !thresholdExceeded {
		emailTemplate = "recipient_notification_positive.txt"
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
			sendEmail(recipient.Email, subject, message)
		}
	}

	if threshold.NotifyUser {
		userEmail := fetchUserEmail(threshold.UserID)
		userMessage := fmt.Sprintf(
			"Hello,\n\nYour MEGGA threshold for %s has been hit. The change was %.2f%%, exceeding your set threshold of %.2f%%.\n\nBest,\nMEGGA",
			dataName, percentChange, threshold.ThresholdValue,
		)
		sendEmail(userEmail, "Your MEGGA Threshold Was Hit - Here's What to Do Next", userMessage)
	}
}