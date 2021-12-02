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

type Handler struct {
	ddbClient DDBAPI

	episodeTableName string
}

func (h *Handler) Handle(ctx context.Context, input events.APIGatewayV2HTTPRequest) (
	*events.APIGatewayV2HTTPResponse, error,
) {
	log.Printf("Request:\n%#v", input)

	filterExp, haveFilter, err := getFilterExpressionFromQueryString(input.QueryStringParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to get filter expression from query, %w", err)
	}

	// Build the DynamoDB expression for retrieving only select fields from the item.
	builder := ddbexp.NewBuilder().
		WithProjection(workshop.ListEpisodesProjection())
	if haveFilter {
		builder.WithFilter(filterExp)
	}
	expr, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression projection, %w", err)
	}

	// Call out to Amazon DynamoDB using Scan to get all the episodes in the table
	episodes, resp, err := h.getEpisodes(ctx, expr)
	if err != nil {
		return nil, fmt.Errorf("get episodes failed, %w", err)
	}
	if resp != nil {
		return resp, nil
	}

	// Respond back with the episode fields selected in the projection.
	return workshop.NewJSONResponse(200, nil, episodes)
}

func (h *Handler) getEpisodes(ctx context.Context, expr ddbexp.Expression) (
	[]workshop.ListEpisodeItem, *events.APIGatewayV2HTTPResponse, error,
) {
	scanPaginator := ddb.NewScanPaginator(h.ddbClient, &ddb.ScanInput{
		TableName:                 &h.episodeTableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	})
	var episodes []workshop.ListEpisodeItem
	for scanPaginator.HasMorePages() {
		result, err := scanPaginator.NextPage(ctx)
		if err != nil {
			resp, err := handleScanError(err)
			return nil, resp, err
		}
		printResponseDebugInformation(result)

		// Convert the DynamoDB AttributeValue datatype into our
		// ListEpisodeItem type.
		pageEpisodes, err := unmarshalEpisodeItems(result.Items)
		if err != nil {
			return nil, nil, err
		}
		episodes = append(episodes, pageEpisodes...)
	}

	return episodes, nil, nil
}

func handleScanError(err error) (*events.APIGatewayV2HTTPResponse, error) {
	var throttleErr *ddbtypes.ProvisionedThroughputExceededException
	if errors.As(err, &throttleErr) {
		log.Printf("Received exception: %v. Returning 429 HTTP Response", err)
		return workshop.NewTooManyRequestsErrorResponse("Please slow down request rate")
	}

	return nil, fmt.Errorf("failed to scan table, %w", err)
}

func unmarshalEpisodeItems(items []map[string]ddbtypes.AttributeValue) (
	[]workshop.ListEpisodeItem, error,
) {
	episodes := make([]workshop.ListEpisodeItem, 0, len(items))
	if err := ddbav.UnmarshalListOfMaps(items, &episodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal episode item, %w", err)
	}
	return episodes, nil
}

func getFilterExpressionFromQueryString(query map[string]string) (
	builder ddbexp.ConditionBuilder, hasCondition bool, err error,
) {
	if v, ok := query["podcast"]; ok && v != "" {
		// Using the expression (aliased as ddbexp) package's
		// ConditionBuilder set the podcastCondition to the condition of the
		// "podcast" attribute equaling the value of "podcast" parameter in
		// the query string.
		return ddbexp.ConditionBuilder{}, false, fmt.Errorf("podcastCondition not implemented")
	}
	if v, ok := query["in-title"]; ok && v != "" {
		// Using the expression (aliased as ddbexp) package's
		// ConditionBuilder set the inTitleCondition to the condition of the
		// "in-title" attribute equaling the value of "in-title" parameter in
		// the query string.
		return ddbexp.ConditionBuilder{}, false, fmt.Errorf("inTitleCondition not implemented")
	}

	return ddbexp.ConditionBuilder{}, false, nil
}

func printResponseDebugInformation(result *ddb.ScanOutput) {
	log.Println("Number of podcasts returned in response: ", len(result.Items))
	if result.LastEvaluatedKey != nil {
		log.Println("Response contains LastEvaluatedKey. There are still more data to be scanned.")
	} else {
		log.Println("Response does not contains LastEvaluatedKey. There is no more data to be scanned.")
	}
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()
	handler := &Handler{
		ddbClient: ddb.NewFromConfig(cfg),

		episodeTableName: envCfg.PodcastEpisodeTableName,
	}

	lambda.Start(handler.Handle)
}

type DDBAPI interface {
	Scan(context.Context, *ddb.ScanInput, ...func(*ddb.Options)) (*ddb.ScanOutput, error)
}
