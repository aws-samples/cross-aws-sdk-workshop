package workshop

import (
	"fmt"
	"strconv"

	ddbexp "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// GetEpisodeByID returns the episode and true if it exists in the list of episodes.
// Otherwise returns false.
func GetEpisodeByID(episodes []Episode, id string) (Episode, bool) {
	for _, e := range episodes {
		if e.ID == id {
			return e, true
		}
	}
	return Episode{}, false
}

// HasEpisodeByID returns true if the episode exists in the list of episodes.
// Otherwise returns false.
func HasEpisodeByID(episodes []Episode, id string) bool {
	for _, e := range episodes {
		if e.ID == id {
			return true
		}
	}
	return false
}

// Episode provides the structure for Episode data passed between Amazon Lambda
// handlers functions in state machines, and the structure for storing the data
// in Amazon DynamoDB.
type Episode struct {
	ID                     string        `json:"id" dynamodbav:"id"`
	Title                  string        `json:"title" dynamodbav:"title"`
	Description            string        `json:"description" dynamodbav:"description"`
	PublishedDate          string        `json:"published" dynamodbav:"published"`
	Podcast                string        `json:"podcast" dynamodbav:"podcast"`
	MediaURL               string        `json:"media_url" dynamodbav:"media_url"`
	MediaContentType       string        `json:"media_content_type" dynamodbav:"media_content_type"`
	MediaKey               string        `json:"media_key" dynamodbav:"media_key"`
	TranscribeExecutionARN string        `json:"transcribe_execution_arn,omitempty" dynamodbav:"transcribe_execution_arn,omitempty"`
	TranscribeJobID        string        `json:"transcribe_job_id,omitempty" dynamodbav:"transcription_job_id,omitempty"`
	TranscribeMetadataKey  string        `json:"transcribe_metadata_key,omitempty" dynamodbav:"transcribe_metadata_key,omitempty"`
	TranscriptionKey       string        `json:"transcription_key,omitempty" dynamodbav:"transcription_key,omitempty"`
	Status                 EpisodeStatus `json:"status" dynamodbav:"status"`
}

// AttributeValuePrimaryKey returns the DynamoDB key for the episode.
func (e Episode) AttributeValuePrimaryKey() map[string]ddbtypes.AttributeValue {
	return map[string]ddbtypes.AttributeValue{
		"id": &ddbtypes.AttributeValueMemberS{Value: e.ID},
	}
}

// ListEpisodeItem provides a type for public fields when listing episodes.
type ListEpisodeItem struct {
	ID      string `json:"id" dynamodbav:"id"`
	Title   string `json:"title" dynamodbav:"title"`
	Podcast string `json:"podcast" dynamodbav:"podcast"`
}

// ListEpisodesProjection returns a DynamoDB expression Projection builder
// for fields to be returned for a episode list response.
func ListEpisodesProjection() ddbexp.ProjectionBuilder {
	return ddbexp.NamesList(
		ddbexp.Name("id"),
		ddbexp.Name("title"),
		ddbexp.Name("podcast"),
	)
}

// DescribeEpisode provides a type for public fields of an Episode.
type DescribeEpisode struct {
	ID          string        `json:"id" dynamodbav:"id"`
	Title       string        `json:"title" dynamodbav:"title"`
	Description string        `json:"description" dynamodbav:"description"`
	Podcast     string        `json:"podcast" dynamodbav:"podcast"`
	Status      EpisodeStatus `json:"status" dynamodbav:"status"`
}

// DescribeEpisodeProjection returns a DynamoDB expression Projection builder
// for fields to be returned for a describe episode response.
func DescribeEpisodeProjection() ddbexp.ProjectionBuilder {
	return ddbexp.NamesList(
		ddbexp.Name("id"),
		ddbexp.Name("title"),
		ddbexp.Name("description"),
		ddbexp.Name("podcast"),
		ddbexp.Name("status"),
	)
}

// EpisodeStatus provides the enumeration of episode statuses.
type EpisodeStatus string

const (
	EpisodeStatusUnknown      EpisodeStatus = ""
	EpisodeStatusPending      EpisodeStatus = "pending"
	EpisodeStatusUploading    EpisodeStatus = "uploading"
	EpisodeStatusTranscribing EpisodeStatus = "transcribing"
	EpisodeStatusProcessing   EpisodeStatus = "processing"
	EpisodeStatusComplete     EpisodeStatus = "complete"
	EpisodeStatusFailure      EpisodeStatus = "failure"
)

func (e EpisodeStatus) String() string { return string(e) }

func (e *EpisodeStatus) UnmarshalJSON(b []byte) error {
	v, err := strconv.Unquote(string(b))
	if err != nil {
		return fmt.Errorf("failed to unquote EpisodeStatus, %w", err)
	}
	ee, err := parseEpisodeStatus(v)
	if err != nil {
		return err
	}
	*e = ee
	return nil
}
func (e EpisodeStatus) MarshalJSON() ([]byte, error) {
	return []byte("\"" + e.String() + "\""), nil
}
func (e *EpisodeStatus) UnmarshalDynamoDBAttributeValue(av ddbtypes.AttributeValue) error {
	avS, ok := av.(*ddbtypes.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("expect string attribute value for episode status, got %T, %v", av, av)
	}

	ee, err := parseEpisodeStatus(avS.Value)
	if err != nil {
		return err
	}
	*e = ee
	return nil
}
func (e EpisodeStatus) MarshalDynamoDBAttributeValue() (ddbtypes.AttributeValue, error) {
	return &ddbtypes.AttributeValueMemberS{Value: e.String()}, nil
}

func parseEpisodeStatus(v string) (EpisodeStatus, error) {
	switch EpisodeStatus(v) {
	case EpisodeStatusPending:
		return EpisodeStatusPending, nil

	case EpisodeStatusUploading:
		return EpisodeStatusUploading, nil

	case EpisodeStatusTranscribing:
		return EpisodeStatusTranscribing, nil

	case EpisodeStatusProcessing:
		return EpisodeStatusProcessing, nil

	case EpisodeStatusComplete:
		return EpisodeStatusComplete, nil

	case EpisodeStatusFailure:
		return EpisodeStatusFailure, nil

	default:
		return EpisodeStatusUnknown, nil
	}
}

func MakeEpisodeRawMediaPath(prefix, episodeID string) string {
	return makeEpisodePrefixPath(prefix, episodeID) + "raw-media"
}

func MakeEpisodeTranscribeMetadataPath(prefix, episodeID string) string {
	return makeEpisodePrefixPath(prefix, episodeID) + "transcribe-metadata.json"
}

func MakeEpisodeTranscriptionPath(prefix, episodeID string) string {
	return makeEpisodePrefixPath(prefix, episodeID) + "transcription.txt"
}

// makeEpisodePrefixPath returns the object prefix path a resource for an
// episode should be stored within an Amazon S3 bucket at.
func makeEpisodePrefixPath(prefix, episodeID string) string {
	return prefix + episodeID + "/"
}

// TranscribeStateMachineInput provides the output structure for Amazon Lambda handlers
// use common input parameters for transcribe of podcast episode.
type TranscribeStateMachineInput struct {
	Episode Episode `json:"episode"`
}

// TranscribeStateMachineOutput provides the output structure for Amazon Lambda handlers
// use common input parameters for transcribe of podcast episode.
type TranscribeStateMachineOutput struct {
	Episode Episode `json:"episode"`
}
