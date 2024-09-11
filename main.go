package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"

	"vjgdienynas/schedule"
)

func main() {
	lambda.StartWithOptions(BuildHandler())
}

func BuildHandler() func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	err := schedule.DefaultDownloader.CheckConnection()
	if err != nil {
		println("schedule downloader error:", err.Error())
	} else {
		println("schedule downloader check passed")
	}
	s := BuildServer()
	adapter := gorillamux.NewV2(s)
	return adapter.ProxyWithContext
}
