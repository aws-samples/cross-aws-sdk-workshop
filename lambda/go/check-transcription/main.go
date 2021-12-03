package main

import (
	"context"
	"fmt"
	"log"

	workshop "aws-workshop"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	tr "github.com/aws/aws-sdk-go-v2/service/transcribe"
)

type Handler struct {
	trClient TranscribeAPI
}

type InputEvent struct {
	Episode workshop.Episode `json:"episode"`
}

type OutputEvent struct {
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason,omitempty"`
}

func (h *Handler) Handle(ctx context.Context, input workshop.TranscribeStateMachineInput) (
	OutputEvent, error,
) {
	log.Println("checking transcription,", input)

	resp, err := h.trClient.GetTranscriptionJob(ctx,
		&tr.GetTranscriptionJobInput{
			TranscriptionJobName: &input.Episode.TranscribeJobID,
		},
	)
	if err != nil {
		return OutputEvent{}, fmt.Errorf("failed to check transcription job, %w", err)
	}

	output := OutputEvent{
		Status:        string(resp.TranscriptionJob.TranscriptionJobStatus),
		FailureReason: aws.ToString(resp.TranscriptionJob.FailureReason),
	}
	log.Println("transcription job status:", output.Status,
		"failure reason:", output.FailureReason)

	return output, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load config, %v", err)
	}

	handler := &Handler{
		trClient: tr.NewFromConfig(cfg),
	}

	lambda.Start(handler.Handle)
}

type TranscribeAPI interface {
	GetTranscriptionJob(context.Context, *tr.GetTranscriptionJobInput, ...func(*tr.Options)) (*tr.GetTranscriptionJobOutput, error)
}
