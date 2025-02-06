package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// AllowedOrigins restricts CORS to only dev and prod frontends
var AllowedOrigins = map[string]bool{
	"http://localhost:5173":                     true, // Dev frontend
	"https://main.dx8te9t0umpd8.amplifyapp.com": true, // Prod frontend
}

// Response structure
type Response struct {
	Message string `json:"message"`
}

// Lambda handler function
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("✅ Lambda function invoked!")

	// Extract the request origin
	origin := request.Headers["origin"]
	if !AllowedOrigins[origin] {
		log.Printf("❌ Unauthorized CORS request from: %s", origin)
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Body:       `{"error":"CORS policy does not allow this origin"}`,
		}, nil
	}

	// Successful response
	response := Response{Message: "✅ Backend connected successfully!"}
	responseBody, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  origin, // ✅ Only allow valid origins
			"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	log.Println("🔗 Initializing AWS Lambda function...")
	lambda.Start(handler)
}
