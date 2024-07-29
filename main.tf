provider "aws" {
  region = var.aws_region
}

resource "aws_iam_role" "lambda_role" {
  name = "lambda_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })

  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
    "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
  ]
}

resource "aws_iam_policy" "appsync_policy" {
  name        = "AppSyncPolicy"
  description = "Policy to allow Lambda to invoke AppSync endpoint"
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "appsync:GraphQL"
        ]
        Resource = [
          var.arn_appsync_api
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_appsync_policy" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.appsync_policy.arn
}

resource "aws_cloudwatch_log_group" "lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.stripe_webhook.function_name}"
  retention_in_days = 1
}

resource "aws_lambda_function" "stripe_webhook" {
  filename         = "lambda-handler.zip"
  function_name    = var.lambda_function_name
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"
  source_code_hash = filebase64sha256("lambda-handler.zip")
  runtime          = "provided.al2023"
  environment {
    variables = {
      STRIPE_WEBHOOK_SECRET = var.stripe_webhook_secret
      GRAPHQL_ENDPOINT      = var.graphql_endpoint
    }
  }
}

resource "aws_lambda_function_url" "stripe_webhook_url" {
  function_name       = aws_lambda_function.stripe_webhook.function_name
  authorization_type  = "NONE"
}

resource "aws_lambda_permission" "allow_public_access" {
  statement_id  = "AllowPublicInvoke"
  action        = "lambda:InvokeFunctionUrl"
  function_name = aws_lambda_function.stripe_webhook.function_name
  principal     = "*"
  function_url_auth_type = aws_lambda_function_url.stripe_webhook_url.authorization_type
}

output "lambda_function_url" {
  value = aws_lambda_function_url.stripe_webhook_url.function_url
}
