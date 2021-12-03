package com.amazonaws.workshop;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayV2HTTPEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayV2HTTPResponse;
import software.amazon.awssdk.core.waiters.WaiterResponse;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.GetObjectRequest;
import software.amazon.awssdk.services.s3.model.HeadObjectResponse;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;
import software.amazon.awssdk.services.s3.presigner.model.GetObjectPresignRequest;
import software.amazon.awssdk.services.s3.presigner.model.PresignedGetObjectRequest;
import software.amazon.awssdk.services.s3.waiters.S3Waiter;

import java.time.Duration;
import java.util.HashMap;
import java.util.Map;

public class PlayPodcast implements RequestHandler<APIGatewayV2HTTPEvent, APIGatewayV2HTTPResponse> {

    private static final String EPISODE_CONTENT_KIND_MEDIA = "media";
    private static final String EPISODE_CONTENT_KIND_TEXT = "text";

    private static final String PODCAST_DATA_KEY_PREFIX = System.getenv("AWS_SDK_WORKSHOP_PODCAST_DATA_KEY_PREFIX");
    private static final String PODCAST_BUCKET = System.getenv("AWS_SDK_WORKSHOP_PODCAST_DATA_BUCKET_NAME");
    private static final String AWS_REGION = System.getenv("AWS_REGION");

    private final S3Client s3;
    private final S3Waiter s3Waiter;
    private final S3Presigner s3Presigner;

    public PlayPodcast() {
        this.s3 = S3Client.create();
        this.s3Waiter = s3.waiter();
        this.s3Presigner = S3Presigner.create();
    }

    @Override
    public APIGatewayV2HTTPResponse handleRequest(APIGatewayV2HTTPEvent event, Context context) {
        String episodeId = event.getPathParameters().get("id");
        String contentType = parseEpisodeContentKind(event.getQueryStringParameters());
        String mediaKey = getEpisodeMediaKey(contentType, episodeId);

        checkMediaExists(mediaKey);

        Map<String, String> headers = new HashMap<>();
        headers.put("location", responseRedirect(PODCAST_BUCKET, mediaKey));
        return APIGatewayV2HTTPResponse.builder().withStatusCode(307).withHeaders(headers).build();
    }

    private String parseEpisodeContentKind(Map<String, String> params) {
        if (params == null){
            return EPISODE_CONTENT_KIND_MEDIA;
        } if (EPISODE_CONTENT_KIND_MEDIA.equals(params.get("content"))){
            return EPISODE_CONTENT_KIND_MEDIA;
        } else if (params.get("content") == "text") {
            return EPISODE_CONTENT_KIND_TEXT;
        }

        throw new IllegalArgumentException("Unable to parse episode content kind!");
    }

    private String getEpisodeMediaKey(String contentType, String episodeId) {
        if (contentType == EPISODE_CONTENT_KIND_MEDIA) {
            return String.format("%s%s/raw-media", PODCAST_DATA_KEY_PREFIX, episodeId);
        } else if (contentType == EPISODE_CONTENT_KIND_TEXT) {
            return String.format("%s%stranscription.txt", PODCAST_DATA_KEY_PREFIX, episodeId);
        }

        throw new RuntimeException("Unable to determine media key!");
    }

    private void checkMediaExists(String mediaKey) {
        WaiterResponse<HeadObjectResponse> objectExistsWaiter = s3Waiter.waitUntilObjectExists(r -> r.key(mediaKey).bucket(PODCAST_BUCKET));

        objectExistsWaiter.matched().exception().ifPresent(e -> {
            throw new RuntimeException("Failed to wait for object to exist", e);
        });
    }

    private String responseRedirect(String bucket, String key) {
        GetObjectRequest request = GetObjectRequest.builder().bucket(bucket).key(key).build();
        GetObjectPresignRequest presignRequest =
                GetObjectPresignRequest
                        .builder()
                        .getObjectRequest(request)
                        .signatureDuration(Duration.ofHours(24))
                        .build();

        PresignedGetObjectRequest presignedGetObjectRequest = s3Presigner.presignGetObject(presignRequest);

        return presignedGetObjectRequest.url().toString();
    }
}
