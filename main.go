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
)

func handler(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	// Log request method
	fmt.Println("Request method:", request.RequestContext.HTTP.Method)

	// Log request headers
	fmt.Println("Received headers:")
	for key, value := range request.Headers {
		fmt.Printf("%s: %s\n", key, value)
	}

	// Log request body
	fmt.Print("Request body:", request.Body)

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

	// Log the event
	fmt.Println("Received event:", event)

	// Handle the event (for example, "checkout.session.completed")
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			fmt.Println("Error unmarshalling event data:", err)
			return events.LambdaFunctionURLResponse{StatusCode: http.StatusBadRequest}, err
		}

		// Extract email from the session
		email := session.Customer.Email
		email2 := session.CustomerEmail
		fmt.Print("Customer email:", email, "Email2:", email2, "Customer:", session.Customer, "Session:", session)
		if email == "" {
			fmt.Println("Customer email is missing in the event data")
			return events.LambdaFunctionURLResponse{StatusCode: http.StatusBadRequest}, fmt.Errorf("customer email is missing in the event data")
		}

		// Call your GraphQL endpoint to update payment_status
		err = updatePaymentStatus(email)
		if err != nil {
			fmt.Println("Error updating payment status:", err)
			return events.LambdaFunctionURLResponse{StatusCode: http.StatusInternalServerError}, err
		}
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
