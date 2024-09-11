package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
)

func main() {
	lambda.StartWithOptions(BuildHandler())
}

func BuildHandler() func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	s, err := BuildServer()
	if err != nil {
		panic(err)
	}
	adapter := gorillamux.NewV2(s)
	return adapter.ProxyWithContext
}
