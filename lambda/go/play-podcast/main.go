package main

import (
	"context"
	"fmt"
	"log"
	"time"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Prevent import not used compile errors. This will be used in later sections
// of the workshop.
var _ = time.Now

type Handler struct {
	s3PresignClient S3PresignAPI
	s3ObjectWaiter  S3ObjectWaiter
	ddbClient       DDBAPI

	awsRegion        string
	bucketName       string
	mediaKeyPrefix   string
	episodeTableName string
}

func (h *Handler) Handle(ctx context.Context, input events.APIGatewayV2HTTPRequest) (
	*events.APIGatewayV2HTTPResponse, error,
) {
	log.Printf("Request:\n%#v", input)

	// Get Episode ID from HTTP URL Path
	episodeID, ok := input.PathParameters["id"]
	if !ok || episodeID == "" {
		return workshop.NewBadRequestErrorResponse("Episode id not provided")
	}

	// Get content to be returned from query string parameter
	contentType, err := parseEpisodeContentKind(input.QueryStringParameters["content"])
	if !ok || episodeID == "" {
		return workshop.NewBadRequestErrorResponse(err.Error())
	}

	// Get the S3 Object key for the episode and content
	mediaKey, err := h.getEpisodeMediaKey(ctx, contentType, episodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episode %v content %v key, %v",
			episodeID, contentType, err)
	}

	if resp, err := h.checkMediaExists(ctx, mediaKey); err != nil || resp != nil {
		return resp, err
	}

	return h.respondRedirect(ctx, mediaKey)
}

func (h *Handler) respondRedirect(ctx context.Context, mediaKey string) (*events.APIGatewayV2HTTPResponse, error) {
	// TODO use the SDK's PresignClient to create a presigned URL for the
	// GetObject API operation.
	return workshop.NewTemporaryRedirectResponse(
		fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", h.awsRegion, h.bucketName, mediaKey),
	)
}

func (h *Handler) checkMediaExists(ctx context.Context, mediaKey string) (*events.APIGatewayV2HTTPResponse, error) {
	// TODO use the SDK's ObjectExistsWaiter to detect if the mediaKey exists,
	// and also allow the lambda handler to wait a short period of time for the
	// object to be present.
	return nil, nil
}

func (h *Handler) getEpisodeMediaKey(ctx context.Context, kind EpisodeContentKind, episodeID string) (
	string, error,
) {
	var mediaKey string
	switch kind {
	case EpisodeContentKindMedia:
		mediaKey = workshop.MakeEpisodeRawMediaPath(h.mediaKeyPrefix, episodeID)

	case EpisodeContentKindText:
		mediaKey = workshop.MakeEpisodeTranscriptionPath(h.mediaKeyPrefix, episodeID)

	default:
		panic("unknown content type, " + string(kind))
	}

	return mediaKey, nil
}

type EpisodeContentKind string

func (e EpisodeContentKind) String() string { return string(e) }
func parseEpisodeContentKind(v string) (EpisodeContentKind, error) {
	switch v {
	case "", string(EpisodeContentKindMedia):
		return EpisodeContentKindMedia, nil

	case string(EpisodeContentKindText):
		return EpisodeContentKindText, nil

	default:
		return "", fmt.Errorf("Unknown content kind, %v", v)
	}
}

const (
	EpisodeContentKindMedia EpisodeContentKind = "media"
	EpisodeContentKindText  EpisodeContentKind = "text"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()

	s3Client := s3.NewFromConfig(cfg)
	handler := &Handler{
		s3PresignClient: s3.NewPresignClient(s3Client),
		s3ObjectWaiter:  s3.NewObjectExistsWaiter(s3Client),
		ddbClient:       ddb.NewFromConfig(cfg),

		awsRegion:        cfg.Region,
		bucketName:       envCfg.PodcastDataBucketName,
		mediaKeyPrefix:   envCfg.PodcastDataKeyPrefix,
		episodeTableName: envCfg.PodcastEpisodeTableName,
	}

	lambda.Start(handler.Handle)
}

type S3PresignAPI interface {
	PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (
		*v4.PresignedHTTPRequest, error,
	)
}
type S3ObjectWaiter interface {
	Wait(context.Context, *s3.HeadObjectInput, time.Duration, ...func(*s3.ObjectExistsWaiterOptions)) error
}
type DDBAPI interface {
	GetItem(context.Context, *ddb.GetItemInput, ...func(*ddb.Options)) (
		*ddb.GetItemOutput, error,
	)
}
