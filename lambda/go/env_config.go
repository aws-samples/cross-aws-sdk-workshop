package workshop

import (
	"os"
	"strconv"
)

const (
	envKeyPrefix                    = "AWS_SDK_WORKSHOP_"
	envKeyTranscribeStateMachineARN = envKeyPrefix + "TRANSCRIBE_STATEMACHINE_ARN"
	envKeyPodcastEpisodeTableName   = envKeyPrefix + "PODCAST_EPISODE_TABLE_NAME"
	envKeyPodcastDataBucketName     = envKeyPrefix + "PODCAST_DATA_BUCKET_NAME"
	envKeyTranscribeAccessRoleARN   = envKeyPrefix + "TRANSCRIBE_ACCESS_ROLE_ARN"

	envKeyPodcastDataKeyPrefix = envKeyPrefix + "PODCAST_DATA_KEY_PREFIX"
	envKeyMaxNumEpisodeImport  = envKeyPrefix + "MAX_NUM_EPISODE_IMPORT"
)

type EnvConfig struct {
	TranscribeStateMachineARN string
	PodcastEpisodeTableName   string
	PodcastDataBucketName     string
	TranscribeAccessRoleARN   string

	PodcastDataKeyPrefix string
	MaxNumEpisodeImport  int
}

func LoadEnvConfig() EnvConfig {
	maxNumEpisodes, _ := strconv.ParseInt(os.Getenv(envKeyMaxNumEpisodeImport), 10, 64)

	return EnvConfig{
		TranscribeStateMachineARN: os.Getenv(envKeyTranscribeStateMachineARN),
		PodcastEpisodeTableName:   os.Getenv(envKeyPodcastEpisodeTableName),
		PodcastDataBucketName:     os.Getenv(envKeyPodcastDataBucketName),
		TranscribeAccessRoleARN:   os.Getenv(envKeyTranscribeAccessRoleARN),

		PodcastDataKeyPrefix: os.Getenv(envKeyPodcastDataKeyPrefix),
		MaxNumEpisodeImport:  int(maxNumEpisodes),
	}
}
