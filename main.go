package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

var (
	stripeSecret    = os.Getenv("STRIPE_WEBHOOK_SECRET")
	graphqlEndpoint = os.Getenv("GRAPHQL_ENDPOINT")
	apiKey          = os.Getenv("API_KEY")
	userID          = "4d8ccbfd-50d8-4421-85e0-cc1c421d4a08"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Verify webhook signature
	payload := request.Body
	sigHeader := request.Headers["Stripe-Signature"]

	event, err := webhook.ConstructEvent([]byte(payload), sigHeader, stripeSecret)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
	}

	// Handle the event
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, err
		}

		// Call your GraphQL endpoint to update payment_status
		err = updatePaymentStatus(userID)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, err
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, nil
}

func updatePaymentStatus(userID string) error {
	query := fmt.Sprintf(`mutation MyMutation {
        updateUser(id: "%s", payment_status: "PAID") {
            id
            payment_status
        }
    }`, userID)

	jsonData := map[string]string{
		"query": query,
	}
	jsonValue, _ := json.Marshal(jsonData)
	req, err := http.NewRequest("POST", graphqlEndpoint, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update payment status, status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
