package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ddbav "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ddbexp "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/aws/smithy-go/rand"
)

type Handler struct {
	sfnClient SFNAPI
	ddbClient DDBAPI

	maxNumEpisodes            int
	episodeTableName          string
	transcribeStateMachineARN string

	httpClient   HTTPDoer
	uuidProvider UUIDProvider
}

type APIInput struct {
	ImportEpisode *ImportEpisode `json:"import_episode"`
	ImportRSSFeed *ImportRSSFeed `json:"import_rss_feed"`
}

// TODO required validation
type ImportEpisode struct {
	ID          string `json:"id"`           // optional
	Title       string `json:"title"`        // required
	Description string `json:"description"`  // optional
	Podcast     string `json:"podcast"`      // optional
	URL         string `json:"url"`          // required
	ContentType string `json:"content_type"` // required if not obtainable via download
}

type ImportRSSFeed struct {
	Title          string `json:"title,omitempty"`  // required
	URL            string `json:"url,omitempty"`    // required
	MaxNumEpisodes int    `json:"max_num_episodes"` // optional
}

type APIOutput struct {
	Episodes []workshop.Episode `json:"episodes"`
}

func (h *Handler) Handle(ctx context.Context, input events.APIGatewayV2HTTPRequest) (
	*events.APIGatewayV2HTTPResponse, error,
) {
	log.Printf("Request:\n%#v", input)

	var apiInput APIInput
	if err := json.Unmarshal([]byte(input.Body), &apiInput); err != nil {
		log.Printf("ERROR: failed to unmarshal request body, %v", err)
		return workshop.NewBadRequestErrorResponse("invalid add podcast request body")
	}

	var episodes []workshop.Episode
	if apiInput.ImportEpisode != nil {
		episode, err := h.importEpisode(apiInput.ImportEpisode)
		if err != nil {
			return nil, fmt.Errorf("failed to import episode, %w", err)
		}
		episodes = append(episodes, episode)
	}

	if apiInput.ImportRSSFeed != nil {
		es, err := h.importRSSFeed(ctx, apiInput.ImportRSSFeed)
		if err != nil {
			return nil, fmt.Errorf("failed to import episodes from RSS feed, %w", err)
		}
		episodes = append(episodes, es...)
	}

	if len(episodes) == 0 {
		return workshop.NewJSONResponse(400, nil, messageOutput{
			Message: "RSS feed did not contain any episodes",
		})
	}

	episodes, err := h.filterEpisodes(ctx, episodes)
	if err != nil {
		return nil, fmt.Errorf("failed to filter episodes, %w", err)
	}
	if len(episodes) == 0 {
		return workshop.NewJSONResponse(200, nil, messageOutput{
			Message: "All episodes in RSS feed up to max number episodes are already imported",
		})
	}

	// Record the episodes
	if err := h.writeEpisodes(ctx, episodes); err != nil {
		return nil, fmt.Errorf("failed to record episodes, %w", err)
	}

	// Kick off imports of episodes
	if err := h.startImport(ctx, episodes); err != nil {
		return nil, fmt.Errorf("failed to import episodes, %w", err)
	}

	return workshop.NewJSONResponse(200, nil, APIOutput{
		Episodes: episodes,
	})
}

type messageOutput struct {
	Message string `json:"message"`
}

func (h *Handler) importEpisode(ep *ImportEpisode) (_ workshop.Episode, err error) {
	log.Printf("importing episode")

	var id string
	if ep.ID != "" {
		id = ep.ID
	} else {
		id, err = makeEpisodeID("", h.uuidProvider)
		if err != nil {
			return workshop.Episode{}, err
		}
	}

	return workshop.Episode{
		ID:               id,
		Title:            ep.Title,
		Description:      ep.Description,
		Podcast:          ep.Podcast,
		MediaURL:         ep.URL,
		MediaContentType: ep.ContentType,
		Status:           workshop.EpisodeStatusPending,
	}, nil
}

func (h *Handler) importRSSFeed(ctx context.Context, feed *ImportRSSFeed) ([]workshop.Episode, error) {
	log.Printf("attempting to import episodes from RSS feed")
	req, err := http.NewRequest("GET", feed.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request for media, %w", err)
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	if _, err = io.Copy(&body, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read RSS feed, %v", err)
	}

	var rss RSS
	if err := xml.NewDecoder(&body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("failed to decode RSS feed, %w", err)
	}

	items := limitItems(rss.Channel.Items, feed.MaxNumEpisodes, h.maxNumEpisodes)
	episodes := make([]workshop.Episode, 0, len(items))
	log.Printf("found %v episodes in RSS", len(items))

	for _, item := range items {
		log.Printf("episode from RSS feed: %#v", item)

		baseID := item.Guid
		if baseID == "" {
			baseID = item.Enclosure.URL
		}

		id, err := makeEpisodeID(baseID, h.uuidProvider)
		if err != nil {
			return nil, err
		}

		episodes = append(episodes, workshop.Episode{
			ID:            id,
			Title:         item.Title,
			Description:   item.Description,
			PublishedDate: item.PublishedDate,
			Podcast:       rss.Channel.Title,
			MediaURL:      item.Enclosure.URL,
			Status:        workshop.EpisodeStatusPending,
		})
	}

	return episodes, nil
}

func (h *Handler) filterEpisodes(ctx context.Context, episodes []workshop.Episode) (
	[]workshop.Episode, error,
) {
	log.Printf("filtering on %v episodes", len(episodes))

	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(episodes))
	for _, episode := range episodes {
		log.Printf("searching for episode, %v", episode.ID)
		keys = append(keys, episode.AttributeValuePrimaryKey())
	}

	unprocessedKeys := map[string]ddbtypes.KeysAndAttributes{
		h.episodeTableName: {
			Keys: keys,
		},
	}

	foundEpisodes := make([]workshop.Episode, 0, len(episodes))
	for len(unprocessedKeys) != 0 {
		resp, err := h.ddbClient.BatchGetItem(ctx, &ddb.BatchGetItemInput{
			RequestItems: unprocessedKeys,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get episodes from DynamoDB, %w", err)
		}
		unprocessedKeys = resp.UnprocessedKeys
		log.Printf("BatchGetItem returned with %v unprocessed items", len(unprocessedKeys))

		foundItems := resp.Responses[h.episodeTableName]
		log.Printf("BatchGetItem returned %v existing episodes", len(foundItems))

		items := make([]workshop.Episode, 0, len(foundItems))
		if err = ddbav.UnmarshalListOfMaps(foundItems, &items); err != nil {
			return nil, fmt.Errorf("failed decode existing episodes in DynamoDB, %w", err)
		}

		foundEpisodes = append(foundEpisodes, items...)
	}

	filteredEpisodes := make([]workshop.Episode, 0, len(episodes))
	for _, episode := range episodes {
		if ep, ok := workshop.GetEpisodeByID(foundEpisodes, episode.ID); ok {
			if ep.Status != workshop.EpisodeStatusFailure {
				log.Printf("filtering out known non failed episode %v", episode.ID)
				continue
			}
		}
		filteredEpisodes = append(filteredEpisodes, episode)
	}

	return filteredEpisodes, nil

}

func (h *Handler) writeEpisodes(ctx context.Context, episodes []workshop.Episode) error {
	writeRequests := make([]ddbtypes.WriteRequest, 0, len(episodes))

	for _, episode := range episodes {
		av, err := ddbav.MarshalMap(episode)
		if err != nil {
			return fmt.Errorf("failed to marshal episodes for DynamoDB, %w", err)
		}

		writeRequests = append(writeRequests, ddbtypes.WriteRequest{
			PutRequest: &ddbtypes.PutRequest{
				Item: av,
			},
		})
	}

	unprocessedItems := map[string][]ddbtypes.WriteRequest{
		h.episodeTableName: writeRequests,
	}

	for len(unprocessedItems) != 0 {
		resp, err := h.ddbClient.BatchWriteItem(ctx, &ddb.BatchWriteItemInput{
			RequestItems: unprocessedItems,
		})
		if err != nil {
			return fmt.Errorf("failed to write episodes to DynamoDB, %w", err)
		}
		unprocessedItems = resp.UnprocessedItems
		log.Printf("BatchWriteItem returned with %v unprocessed items", len(unprocessedItems))
	}

	return nil
}

func (h *Handler) startImport(ctx context.Context, episodes []workshop.Episode) error {
	for i, episode := range episodes {
		if episode.MediaURL == "" {
			log.Printf("skipping episode %v, has no media URL", episode.ID)
			episode.Status = workshop.EpisodeStatusComplete
			// Update execution ARN in for episode in table
			if err := h.updateEpisodeStatus(ctx, episode); err != nil {
				return fmt.Errorf("update %v failed, %v", episode.ID, err)
			}
			continue
		}

		input, err := json.Marshal(workshop.TranscribeStateMachineInput{Episode: episode})

		resp, err := h.sfnClient.StartExecution(ctx, &sfn.StartExecutionInput{
			StateMachineArn: &h.transcribeStateMachineARN,
			Input:           aws.String(string(input)),
		})
		if err != nil {
			return fmt.Errorf("failed to start episode %v transcribe, %w", episode.ID, err)
		}
		executionARN := aws.ToString(resp.ExecutionArn)
		log.Printf("starting transcribe for %v, %v", episode.ID, executionARN)

		episodes[i].TranscribeExecutionARN = executionARN

		// Update execution ARN in for episode in table
		if err := h.updateEpisodeExecutionARN(ctx, episode); err != nil {
			return fmt.Errorf("update %v failed, %v", episode.ID, err)
		}
	}

	return nil
}

func (h *Handler) updateEpisodeStatus(ctx context.Context, episode workshop.Episode) error {
	log.Printf("updating episode %v with status, %v",
		episode.ID, episode.Status)

	exp, err := ddbexp.NewBuilder().WithUpdate(
		ddbexp.Set(
			ddbexp.Name("status"),
			ddbexp.Value(episode.Status),
		),
	).Build()
	if err != nil {
		return fmt.Errorf("failed to build update expression, %w", err)
	}

	_, err = h.ddbClient.UpdateItem(ctx, &ddb.UpdateItemInput{
		TableName: &h.episodeTableName,
		Key: workshop.Episode{
			ID: episode.ID,
		}.AttributeValuePrimaryKey(),
		UpdateExpression:          exp.Update(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
	})
	if err != nil {
		return fmt.Errorf("failed to update episode, %w", err)
	}

	return nil
}

func (h *Handler) updateEpisodeExecutionARN(ctx context.Context, episode workshop.Episode) error {
	log.Printf("updating episode %v with execution ARN, %v",
		episode.ID, episode.TranscribeExecutionARN)

	exp, err := ddbexp.NewBuilder().WithUpdate(
		ddbexp.Set(
			ddbexp.Name("transcribe_execution_arn"),
			ddbexp.Value(episode.TranscribeExecutionARN),
		),
	).Build()
	if err != nil {
		return fmt.Errorf("failed to build update expression, %w", err)
	}

	_, err = h.ddbClient.UpdateItem(ctx, &ddb.UpdateItemInput{
		TableName: &h.episodeTableName,
		Key: workshop.Episode{
			ID: episode.ID,
		}.AttributeValuePrimaryKey(),
		UpdateExpression:          exp.Update(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
	})
	if err != nil {
		return fmt.Errorf("failed to update episode, %w", err)
	}

	return nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()
	handler := &Handler{
		sfnClient: sfn.NewFromConfig(cfg),
		ddbClient: ddb.NewFromConfig(cfg),

		maxNumEpisodes:            envCfg.MaxNumEpisodeImport,
		episodeTableName:          envCfg.PodcastEpisodeTableName,
		transcribeStateMachineARN: envCfg.TranscribeStateMachineARN,

		httpClient:   &http.Client{},
		uuidProvider: rand.NewUUID(rand.Reader),
	}

	lambda.Start(handler.Handle)
}

type SFNAPI interface {
	StartExecution(context.Context, *sfn.StartExecutionInput, ...func(*sfn.Options)) (
		*sfn.StartExecutionOutput, error,
	)
}

type DDBAPI interface {
	BatchWriteItem(context.Context, *ddb.BatchWriteItemInput, ...func(*ddb.Options)) (
		*ddb.BatchWriteItemOutput, error,
	)
	BatchGetItem(context.Context, *ddb.BatchGetItemInput, ...func(*ddb.Options)) (
		*ddb.BatchGetItemOutput, error,
	)
	UpdateItem(context.Context, *ddb.UpdateItemInput, ...func(*ddb.Options)) (
		*ddb.UpdateItemOutput, error,
	)
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type UUIDProvider interface {
	GetUUID() (string, error)
}

func makeEpisodeID(id string, provider UUIDProvider) (_ string, err error) {
	if id != "" {
		// RSS feed GUID are an opaque values, in order to prevent issues with
		// various places the Id is used, create a hash of the original GUID
		// and store that instead.
		h := sha256.New()
		h.Write([]byte(id))
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	id, err = provider.GetUUID()
	if err != nil {
		return "", fmt.Errorf("failed to get new UUID, %w", err)
	}
	return id, nil
}

func limitItems(items []Item, ask, max int) []Item {
	if ask == 0 && max == 0 {
		return items
	}
	if ask == 0 || ask > max {
		ask = max
	}

	if len(items) < ask {
		return items
	}

	return items[:ask]
}
