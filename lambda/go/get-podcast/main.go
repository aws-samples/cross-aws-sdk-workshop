package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	ddbav "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddbexp "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Prevent import not used compile errors. This will be used in later sections
// of the workshop.
var _ ddbtypes.ProvisionedThroughputExceededException
var _ = errors.As

type Handler struct {
	ddbClient DDBAPI

	episodeTableName string
}

func (h *Handler) Handle(ctx context.Context, input events.APIGatewayV2HTTPRequest) (
	*events.APIGatewayV2HTTPResponse, error,
) {
	log.Printf("Request:\n%#v", input)

	episodeID, ok := input.PathParameters["id"]
	if !ok || episodeID == "" {
		return workshop.NewBadRequestErrorResponse("Episode id not provided")
	}

	// Build the DynamoDB expression for retrieving only select fields from the item.
	expr, err := ddbexp.NewBuilder().
		WithProjection(workshop.DescribeEpisodeProjection()).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression projection, %w", err)
	}

	// Call out to Amazon DynamoDB using GetItem API to get the specific
	// episode from the table.
	result, err := h.ddbClient.GetItem(ctx, &ddb.GetItemInput{
		TableName:                &h.episodeTableName,
		Key:                      workshop.Episode{ID: episodeID}.AttributeValuePrimaryKey(),
		ExpressionAttributeNames: expr.Names(),
		ProjectionExpression:     expr.Projection(),
	})
	if err != nil {
		return handleGetItemError(err)
	}
	if len(result.Item) == 0 {
		return workshop.NewNotFoundErrorResponse("Podcast not found")
	}

	// Convert the DynamoDB AttributeValue datatype into our Episode Go type.
	var episode workshop.DescribeEpisode
	if err := ddbav.UnmarshalMap(result.Item, &episode); err != nil {
		return nil, fmt.Errorf("failed to unmarshal episode item, %w", err)
	}

	// Respond back with the episode fields selected in the projection.
	return workshop.NewJSONResponse(200, nil, episode)
}

func handleGetItemError(err error) (*events.APIGatewayV2HTTPResponse, error) {
	var throttleErr *ddbtypes.ProvisionedThroughputExceededException
	if errors.As(err, &throttleErr) {
		log.Printf("Received exception: %v. Returning 429 HTTP Response", err)
		return workshop.NewTooManyRequestsErrorResponse("Please slow down request rate")
	}

	return nil, fmt.Errorf("failed to get item from table, %w", err)
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()
	handler := &Handler{
		ddbClient:        ddb.NewFromConfig(cfg),
		episodeTableName: envCfg.PodcastEpisodeTableName,
	}

	lambda.Start(handler.Handle)
}

type DDBAPI interface {
	GetItem(context.Context, *ddb.GetItemInput, ...func(*ddb.Options)) (
		*ddb.GetItemOutput, error,
	)
}
