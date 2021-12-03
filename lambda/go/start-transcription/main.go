package main

import (
	"context"
	"fmt"
	"log"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	tr "github.com/aws/aws-sdk-go-v2/service/transcribe"
	trtypes "github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/aws/smithy-go/rand"
)

type Handler struct {
	s3EndpointResolver s3.EndpointResolver
	trClient           TranscribeAPI

	region           string
	bucketAccessRole string
	bucketName       string
	mediaKeyPrefix   string

	uuidProvider UUIDProvider
}

func (h *Handler) Handle(ctx context.Context, input workshop.TranscribeStateMachineInput) (
	workshop.TranscribeStateMachineOutput, error,
) {
	log.Println("staring transcription,", input)
	episode := input.Episode

	mediaFormat, err := contentTypeToMediaFormat(episode.MediaContentType)
	if err != nil {
		return workshop.TranscribeStateMachineOutput{}, err
	}

	mediaURI, err := h.getS3Endpoint(episode.MediaKey)
	if err != nil {
		return workshop.TranscribeStateMachineOutput{}, err
	}

	episode.TranscribeMetadataKey = workshop.MakeEpisodeTranscribeMetadataPath(
		h.mediaKeyPrefix, episode.ID,
	)

	if episode.TranscribeJobID != "" {
		log.Println("transcription already started,", episode.TranscribeJobID)
		return workshop.TranscribeStateMachineOutput{Episode: episode}, nil
	}

	episode.TranscribeJobID, err = h.uuidProvider.GetUUID()
	if err != nil {
		return workshop.TranscribeStateMachineOutput{}, err
	}
	resp, err := h.trClient.StartTranscriptionJob(ctx,
		&tr.StartTranscriptionJobInput{
			TranscriptionJobName: &episode.TranscribeJobID,
			IdentifyLanguage:     aws.Bool(true),
			MediaFormat:          mediaFormat,
			Media: &trtypes.Media{
				MediaFileUri: &mediaURI,
			},
			Settings: &trtypes.Settings{
				MaxSpeakerLabels:  aws.Int32(10),
				ShowSpeakerLabels: aws.Bool(true),
			},
			JobExecutionSettings: &trtypes.JobExecutionSettings{
				AllowDeferredExecution: aws.Bool(true),
				DataAccessRoleArn:      &h.bucketAccessRole,
			},
			OutputBucketName: &h.bucketName,
			OutputKey:        &episode.TranscribeMetadataKey,
		},
	)
	if err != nil {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("failed to start transcription job, %w", err)
	}

	log.Println("transcription started,", episode.TranscribeJobID, resp)

	return workshop.TranscribeStateMachineOutput{
		Episode: episode,
	}, nil
}

func (h *Handler) getS3Endpoint(key string) (string, error) {
	endpoint, err := h.s3EndpointResolver.ResolveEndpoint(h.region, s3.EndpointResolverOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get S3 endpoint for region %v, %w", h.region, err)
	}

	// TODO this does not escape bucket or key names.
	return endpoint.URL + "/" + h.bucketName + "/" + key, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()
	handler := &Handler{
		s3EndpointResolver: s3.NewDefaultEndpointResolver(),
		region:             cfg.Region,
		trClient:           tr.NewFromConfig(cfg),
		bucketAccessRole:   envCfg.TranscribeAccessRoleARN,
		bucketName:         envCfg.PodcastDataBucketName,
		mediaKeyPrefix:     envCfg.PodcastDataKeyPrefix,

		uuidProvider: rand.NewUUID(rand.Reader),
	}

	lambda.Start(handler.Handle)
}

type UUIDProvider interface {
	GetUUID() (string, error)
}
type TranscribeAPI interface {
	StartTranscriptionJob(ctx context.Context, params *tr.StartTranscriptionJobInput, optFns ...func(*tr.Options)) (*tr.StartTranscriptionJobOutput, error)
}

func contentTypeToMediaFormat(v string) (trtypes.MediaFormat, error) {
	switch v {
	case "audio/mpeg":
		return trtypes.MediaFormatMp3, nil
	case "audio/wav":
		return trtypes.MediaFormatWav, nil
	case "audio/flac":
		return trtypes.MediaFormatFlac, nil
	case "audio/mp4a-latm":
		return trtypes.MediaFormatMp4, nil
	default:
		return "", fmt.Errorf("unsupported media content type, %v", v)
	}
}
