package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ddbav "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Handler struct {
	s3Uploader   S3UploadAPI
	s3Downloader S3DownloadAPI
	ddbClient    DDBAPI

	bucketName       string
	mediaKeyPrefix   string
	episodeTableName string
}

func (h *Handler) Handle(ctx context.Context, input workshop.TranscribeStateMachineInput) (
	workshop.TranscribeStateMachineOutput, error,
) {
	log.Println("processing transcription,", input)
	episode := input.Episode

	transcribeOutput := manager.NewWriteAtBuffer(make([]byte, 0, 1*1024*1024))
	_, err := h.s3Downloader.Download(ctx, transcribeOutput, &s3.GetObjectInput{
		Bucket: &h.bucketName,
		Key:    &episode.TranscribeMetadataKey,
	})
	if err != nil {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("failed to download transcribe metadata, %w", err)
	}

	var transcribeMetadata = struct {
		Results struct {
			LanguageCode string `json:"language_code"`
			Transcripts  []struct {
				Transcript string `json:"transcript"`
			} `json:"transcripts"`
		} `json:"results"`
	}{}
	if err = json.Unmarshal(transcribeOutput.Bytes(), &transcribeMetadata); err != nil {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("failed to decode transcribe metadata, %w", err)
	}
	if len(transcribeMetadata.Results.Transcripts) == 0 {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("transcribe metadata did not contain transcription, %v", transcribeMetadata)
	}

	// TODO Do some kind of processing on transcription metadata
	var transcriptBuffer bytes.Buffer
	for _, result := range transcribeMetadata.Results.Transcripts {
		transcriptBuffer.WriteString(result.Transcript)
		transcriptBuffer.WriteString("\n\n")
	}

	episode.TranscriptionKey = workshop.MakeEpisodeTranscriptionPath(
		h.mediaKeyPrefix, episode.ID,
	)

	_, err = h.s3Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &h.bucketName,
		Key:         &episode.TranscriptionKey,
		ContentType: aws.String("text/plain"),
		Body:        bytes.NewReader(transcriptBuffer.Bytes()),
	})
	if err != nil {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("failed to upload transcription file, %w", err)
	}
	log.Println("uploaded media transcription,", episode.TranscriptionKey)

	av, err := ddbav.MarshalMap(episode)
	if err != nil {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("failed to marshal episode, %w", err)
	}

	log.Println("updating episode table,", episode)
	_, err = h.ddbClient.PutItem(ctx, &ddb.PutItemInput{
		TableName: &h.episodeTableName,
		Item:      av,
	})
	if err != nil {
		return workshop.TranscribeStateMachineOutput{},
			fmt.Errorf("failed to put episode to metadata table, %w", err)
	}

	return workshop.TranscribeStateMachineOutput{
		Episode: input.Episode,
	}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()

	s3Client := s3.NewFromConfig(cfg)
	handler := &Handler{
		s3Uploader:   manager.NewUploader(s3Client),
		s3Downloader: manager.NewDownloader(s3Client),
		ddbClient:    ddb.NewFromConfig(cfg),

		bucketName:       envCfg.PodcastDataBucketName,
		mediaKeyPrefix:   envCfg.PodcastDataKeyPrefix,
		episodeTableName: envCfg.PodcastEpisodeTableName,
	}

	lambda.Start(handler.Handle)
}

type S3UploadAPI interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}
type S3DownloadAPI interface {
	Download(context.Context, io.WriterAt, *s3.GetObjectInput, ...func(*manager.Downloader)) (int64, error)
}
type DDBAPI interface {
	PutItem(context.Context, *ddb.PutItemInput, ...func(*ddb.Options)) (*ddb.PutItemOutput, error)
}
