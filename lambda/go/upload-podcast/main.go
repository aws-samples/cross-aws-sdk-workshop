package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Handler struct {
	httpClient HTTPDoer
	s3Uploader S3UploadAPI

	bucketName     string
	mediaKeyPrefix string
}

func (h *Handler) Handle(ctx context.Context, input workshop.TranscribeStateMachineInput) (
	*workshop.TranscribeStateMachineOutput, error,
) {
	log.Println("staring upload,", input)
	log.Printf("Downloading podcast from: %v", input.Episode.MediaURL)
	episode := input.Episode

	req, err := http.NewRequest("GET", episode.MediaURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request for media, %w", err)
	}
	defer resp.Body.Close()
	respBody := resp.Body

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to get media %v", resp.StatusCode)
	}

	// Get content type if not already specified on the episode.
	if episode.MediaContentType == "" || episode.MediaContentType == "application/octet-stream" {
		episode.MediaContentType = resp.Header.Get("Content-Type")
	}
	// Guess at content type by content of the file.
	if episode.MediaContentType == "" {
		var buf bytes.Buffer
		n, err := io.Copy(&buf, io.LimitReader(resp.Body, 512))
		if err != nil {
			return nil, fmt.Errorf("failed to read media to detect content-type %w", err)
		}
		episode.MediaContentType = http.DetectContentType(buf.Bytes()[:n])

		// Wrap buffered bytes and remaining response body together to be
		// uploaded together.
		respBody = ioutil.NopCloser(io.MultiReader(&buf, resp.Body))
	}
	_ = respBody

	episode.MediaKey = workshop.MakeEpisodeRawMediaPath(h.mediaKeyPrefix, episode.ID)

	if err := h.uploadMedia(ctx, episode.MediaKey, episode.MediaContentType, respBody); err != nil {
		return nil, fmt.Errorf("upload episode media failed, %v", err)
	}

	return &workshop.TranscribeStateMachineOutput{
		Episode: episode,
	}, nil
}

func (h *Handler) uploadMedia(ctx context.Context, mediaKey, mediaContentType string, mediaContent io.Reader) error {
	_, err := h.s3Uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &h.bucketName,
		Key:         &mediaKey,
		ContentType: &mediaContentType,
		Body:        mediaContent,
	})
	if err != nil {
		return err
	}
	log.Printf("uploaded episode, s3://%s/%s", h.bucketName, mediaKey)

	return nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	envCfg := workshop.LoadEnvConfig()
	handler := &Handler{
		httpClient:     &http.Client{},
		s3Uploader:     manager.NewUploader(s3.NewFromConfig(cfg)),
		bucketName:     envCfg.PodcastDataBucketName,
		mediaKeyPrefix: envCfg.PodcastDataKeyPrefix,
	}

	lambda.Start(handler.Handle)
}

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}
type S3UploadAPI interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}
