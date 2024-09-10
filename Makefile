build-LambdaHandler:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap .

	cp bootstrap $(ARTIFACTS_DIR)/bootstrap


lint:
	sam validate --lint

server:
	sam local start-api