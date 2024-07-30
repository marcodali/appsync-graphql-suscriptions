package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

var (
	stripeSecret    = os.Getenv("STRIPE_WEBHOOK_SECRET")
	graphqlEndpoint = os.Getenv("GRAPHQL_ENDPOINT")
	apiKey          = os.Getenv("API_KEY")
)

func debugStripeEvent(event stripe.Event) {
	// Parse event object as JSON and display its properties
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling event:", err)
		return
	}

	// Remove new lines and print the JSON data in one line
	oneLineData := strings.ReplaceAll(string(data), "\n", " ")
	fmt.Println("Event data:", oneLineData)
}

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Verify webhook signature
	payload := request.Body
	sigHeader, ok := request.Headers["stripe-signature"]
	if !ok {
		fmt.Println("Missing stripe-signature header")
		return events.LambdaFunctionURLResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Missing stripe-signature header",
		}, fmt.Errorf("webhook has no stripe-signature header")
	}

	event, err := webhook.ConstructEvent([]byte(payload), sigHeader, stripeSecret)
	if err != nil {
		fmt.Println("Error constructing webhook event:", err)
		return events.LambdaFunctionURLResponse{StatusCode: http.StatusBadRequest}, err
	}

	// debugStripeEvent(event)

	if event.Type != "checkout.session.completed" {
		fmt.Println("Unhandled event type:", event.Type)
		return events.LambdaFunctionURLResponse{StatusCode: http.StatusOK}, nil
	}

	// Call your GraphQL endpoint to update payment_status
	email := event.GetObjectValue("customer_details", "email")
	err = updatePaymentStatus(email)
	if err != nil {
		fmt.Println("Error updating payment status:", err)
		return events.LambdaFunctionURLResponse{StatusCode: http.StatusInternalServerError}, err
	}

	return events.LambdaFunctionURLResponse{StatusCode: http.StatusOK}, nil
}

func updatePaymentStatus(email string) error {
	query := fmt.Sprintf(`mutation MyMutation {
        updateUser(id: "%s", payment_status: "PAID") {
            id
            payment_status
        }
    }`, email)

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
	fmt.Println("func ver", "gacela")
	lambda.Start(handler)
}
