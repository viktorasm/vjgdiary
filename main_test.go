package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	r := require.New(t)
	if os.Getenv("E2E_USER") == "" {
		t.Skip("skipping e2e test")
	}

	user := os.Getenv("E2E_USER")
	password := os.Getenv("E2E_PASSWORD")

	handler := BuildHandler()

	result, err := handler(context.Background(), events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/lesson-info",
		RawQueryString: "",
		Cookies:        nil,
		Headers: map[string]string{
			"METHOD": "POST",
		},
		QueryStringParameters: nil,
		PathParameters:        nil,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: "POST",
			},
		},
		StageVariables: nil,
		Body: toJSON(t, &LoginRequest{
			Username: user,
			Password: password,
		}),
		IsBase64Encoded: false,
	})
	r.NoError(err)
	r.Equal(http.StatusOK, result.StatusCode)

	println(result.Body)
}

func toJSON(t testing.TB, v interface{}) string {
	t.Helper()
	result, err := json.Marshal(v)
	require.NoError(t, err)
	return string(result)
}
