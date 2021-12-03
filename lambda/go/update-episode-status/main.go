package main

import (
	"context"
	"fmt"
	"log"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	ddbexp "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Handler struct {
	ddbClient DDBAPI

	episodeTableName string
}

type InputEvent struct {
	EpisodeID string `json:"id"`
	Status    string `json:"status"`
}

func (h *Handler) Handle(ctx context.Context, input InputEvent) (
	string, error,
) {
	log.Println("updating episode status,", input)

	exp, err := ddbexp.NewBuilder().WithUpdate(
		ddbexp.Set(
			ddbexp.Name("status"),
			ddbexp.Value(input.Status),
		),
	).Build()
	if err != nil {
		return "", fmt.Errorf("failed to build update expression, %w", err)
	}

	_, err = h.ddbClient.UpdateItem(ctx, &ddb.UpdateItemInput{
		TableName: &h.episodeTableName,
		Key: workshop.Episode{
			ID: input.EpisodeID,
		}.AttributeValuePrimaryKey(),
		UpdateExpression:          exp.Update(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to update episode, %w", err)
	}

	log.Printf("episode %v updated, %v", input.EpisodeID, input.Status)

	return input.Status, nil
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
	UpdateItem(context.Context, *ddb.UpdateItemInput, ...func(*ddb.Options)) (
		*ddb.UpdateItemOutput, error,
	)
}
