package main

import (
	"context"
	"encoding/json"
	"log"
	"megga-backend/internal/database"
	"megga-backend/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response structure for API response
type Response struct {
	Message string `json:"message"`
}

// Lambda handler function
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("✅ Lambda function invoked!")

	// Fetch latest BLS data using the service function
	err := services.FetchLatestBLSData(database.DB)
	if err != nil {
		log.Printf("❌ Error fetching BLS data: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error":"Failed to fetch BLS data"}`,
		}, nil
	}

	log.Println("✅ Successfully updated BLS data!")

	// Return success response
	response := Response{Message: "✅ BLS data updated successfully!"}
	responseBody, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	log.Println("🔗 Initializing AWS Lambda function for BLS data updates...")
	database.InitDB()
	defer database.CloseDB()

	lambda.Start(handler)
}
