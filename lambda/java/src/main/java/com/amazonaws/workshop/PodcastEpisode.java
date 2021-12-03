package com.amazonaws.workshop;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbAttribute;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbBean;
import software.amazon.awssdk.enhanced.dynamodb.mapper.annotations.DynamoDbPartitionKey;

/**
 * {
 *   "id": "REPLACE",
 *   "media_content_type": "audio/mpeg",
 *   "media_key": "podcasts/1234-5678-980/raw-media.mp3",
 *   "media_url": "https://d1le29qyzha1u4.cloudfront.net/AWS_Podcast_Episode_478.mp3",
 *   "title": "A sample title for the podcast episode",
 *   "description": "A sample description for the podcast episode",
 *   "podcast": "AWS Podcast",
 *   "status": "complete",
 *   "published": "Wed, 13 Oct 2021 19:32:37 GMT",
 *   "transcribe_metadata_key": "podcasts/1234-5678-980/transcribe-metadata.json",
 *   "transcription_job_id": "9e6b4895-8385-4e59-9f96-7b0df00cfac6",
 *   "transcription_key": "podcasts/1234-5678-980/transcription.txt"
 * }
 */
@DynamoDbBean
@JsonInclude(JsonInclude.Include.NON_NULL)
public class PodcastEpisode {
    private String id;
    private String mediaContentType;
    private String mediaKey;
    private String mediaUrl;
    private String title;
    private String description;
    private String podcast;
    private String status;
    private String published;
    private String transcribeMetadataKey;
    private String transcriptionJobId;
    private String transcriptionKey;

    public PodcastEpisode() {}

    @DynamoDbPartitionKey
    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    @DynamoDbAttribute("media_content_type")
    @JsonProperty("media_content_type")
    public String getMediaContentType() {
        return mediaContentType;
    }

    @JsonProperty("media_content_type")
    public void setMediaContentType(String mediaContentType) {
        this.mediaContentType = mediaContentType;
    }

    @DynamoDbAttribute("media_key")
    @JsonProperty("media_key")
    public String getMediaKey() {
        return mediaKey;
    }

    @JsonProperty("media_key")
    public void setMediaKey(String mediaKey) {
        this.mediaKey = mediaKey;
    }

    @DynamoDbAttribute("media_url")
    @JsonProperty("media_url")
    public String getMediaUrl() {
        return mediaUrl;
    }

    @JsonProperty("media_url")
    public void setMediaUrl(String mediaUrl) {
        this.mediaUrl = mediaUrl;
    }

    public String getTitle() {
        return title;
    }

    public void setTitle(String title) {
        this.title = title;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getPodcast() {
        return podcast;
    }

    public void setPodcast(String podcast) {
        this.podcast = podcast;
    }

    public String getStatus() {
        return status;
    }

    public void setStatus(String status) {
        this.status = status;
    }

    public String getPublished() {
        return published;
    }

    public void setPublished(String published) {
        this.published = published;
    }

    @DynamoDbAttribute("transcribe_metadata_key")
    @JsonProperty("transcribe_metadata_key")
    public String getTranscribeMetadataKey() {
        return transcribeMetadataKey;
    }

    @JsonProperty("transcribe_metadata_key")
    public void setTranscribeMetadataKey(String transcribeMetadataKey) {
        this.transcribeMetadataKey = transcribeMetadataKey;
    }

    @DynamoDbAttribute("transcription_job_id")
    @JsonProperty("transcription_job_id")
    public String getTranscriptionJobId() {
        return transcriptionJobId;
    }

    @JsonProperty("transcription_job_id")
    public void setTranscriptionJobId(String transcriptionJobId) {
        this.transcriptionJobId = transcriptionJobId;
    }

    @DynamoDbAttribute("transcription_key")
    @JsonProperty("transcription_key")
    public String getTranscriptionKey() {
        return transcriptionKey;
    }

    public void setTranscriptionKey(String transcriptionKey) {
        this.transcriptionKey = transcriptionKey;
    }
}
