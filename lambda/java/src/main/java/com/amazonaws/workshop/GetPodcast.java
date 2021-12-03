package com.amazonaws.workshop;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayV2HTTPEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayV2HTTPResponse;
import software.amazon.awssdk.core.retry.RetryPolicy;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbEnhancedClient;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import software.amazon.awssdk.enhanced.dynamodb.Key;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.ProvisionedThroughputExceededException;

public class GetPodcast implements RequestHandler<APIGatewayV2HTTPEvent, APIGatewayV2HTTPResponse> {

    private static final String PODCAST_TABLE = System.getenv("AWS_SDK_WORKSHOP_PODCAST_EPISODE_TABLE_NAME");

    private final DynamoDbClient dynamoDbClient;
    private final DynamoDbEnhancedClient enhancedClient;

    public GetPodcast() {
        this.dynamoDbClient = DynamoDbClient
                .builder()
                .overrideConfiguration(o -> o.retryPolicy(RetryPolicy.defaultRetryPolicy().toBuilder().numRetries(3).build()))
                .build();

        this.enhancedClient = DynamoDbEnhancedClient
                .builder()
                .dynamoDbClient(dynamoDbClient)
                .build();
    }

    @Override
    public APIGatewayV2HTTPResponse handleRequest(APIGatewayV2HTTPEvent event, Context context) {
        DynamoDbTable<PodcastEpisode> podcastTable = enhancedClient.table(PODCAST_TABLE, WorkshopUtils.GET_PODCAST_TABLE_SCHEMA);

        String episodeId = event.getPathParameters().get("id");
        if (episodeId == null || episodeId == "") {
            throw new IllegalArgumentException("Episode Id not provided");
        }

        Key key = Key.builder().partitionValue(event.getPathParameters().get("id")).build();

        PodcastEpisode podcastEpisode = podcastTable.getItem(key);
        try {
            return APIGatewayV2HTTPResponse.builder().withStatusCode(200).withBody(WorkshopUtils.writeValue(podcastEpisode)).build();
        } catch (Exception e) {
            throw new RuntimeException("Unable to get podcast", e);
        }
    }
}
