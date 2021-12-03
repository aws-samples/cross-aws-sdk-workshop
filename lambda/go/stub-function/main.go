package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3API interface {
}

type DDBAPI interface {
}

type Handler struct {
	s3Client  S3API
	ddbClient DDBAPI
}

func (h *Handler) Handle(ctx context.Context, input events.APIGatewayV2HTTPRequest) (
	events.APIGatewayV2HTTPResponse, error,
) {

	return events.APIGatewayV2HTTPResponse{}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	handler := &Handler{
		s3Client:  s3.NewFromConfig(cfg),
		ddbClient: ddb.NewFromConfig(cfg),
	}

	lambda.Start(handler.Handle)
}
