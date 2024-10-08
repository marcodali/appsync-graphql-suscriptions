variable "aws_region" {
  description = "The AWS region to deploy to"
}

variable "lambda_function_name" {
  description = "The name of the Lambda function"
}

variable "stripe_webhook_secret" {
  description = "Stripe webhook secret"
}

variable "graphql_endpoint" {
  description = "GraphQL endpoint to update payment status"
}

variable "api_key" {
  description = "API key for accessing the AppSync endpoint"
}
