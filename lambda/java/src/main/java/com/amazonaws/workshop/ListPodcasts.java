package com.amazonaws.workshop;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayV2HTTPEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayV2HTTPResponse;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbEnhancedClient;
import software.amazon.awssdk.enhanced.dynamodb.DynamoDbTable;
import software.amazon.awssdk.enhanced.dynamodb.Expression;
import software.amazon.awssdk.enhanced.dynamodb.model.PageIterable;
import software.amazon.awssdk.enhanced.dynamodb.model.ScanEnhancedRequest;
import software.amazon.awssdk.services.dynamodb.DynamoDbClient;
import software.amazon.awssdk.services.dynamodb.model.AttributeValue;
import software.amazon.awssdk.services.dynamodb.model.ScanRequest;
import software.amazon.awssdk.services.dynamodb.model.ScanResponse;
import software.amazon.awssdk.utils.ImmutableMap;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public class ListPodcasts implements RequestHandler<APIGatewayV2HTTPEvent, APIGatewayV2HTTPResponse> {

    private static final String PODCAST_TABLE = System.getenv("AWS_SDK_WORKSHOP_PODCAST_EPISODE_TABLE_NAME");

    private final DynamoDbClient dynamoDbClient;
    private final DynamoDbEnhancedClient enhancedClient;

    public ListPodcasts() {
        this.dynamoDbClient = DynamoDbClient.create();

        this.enhancedClient = DynamoDbEnhancedClient
                .builder()
                .dynamoDbClient(dynamoDbClient)
                .build();
    }

    @Override
    public APIGatewayV2HTTPResponse handleRequest(APIGatewayV2HTTPEvent event, Context context) {
        DynamoDbTable<PodcastEpisode> podcastTable = enhancedClient.table(PODCAST_TABLE, WorkshopUtils.LIST_PODCAST_TABLE_SCHEMA);

        ScanRequest.Builder scanRequest =
                ScanRequest
                        .builder()
                        .tableName(PODCAST_TABLE)
                        .projectionExpression("id, #t, #p")
                        .expressionAttributeNames(ImmutableMap.of("#t", "title", "#p", "podcast"));

        if (event.getQueryStringParameters() != null) {
            scanRequest.filterExpression(getFilterExpressionFromQueryParams(event.getQueryStringParameters()).expression());
        }

        ScanResponse response = dynamoDbClient.scan(scanRequest.build());

        System.out.println(String.format("Number of podcasts returned in response: %d", response.count()));

        if (response.hasLastEvaluatedKey()) {
            System.out.println("Response contains LastEvaluatedKey. There is still more data to be scanned.");
        } else {
            System.out.println("Response does not contain LastEvaluatedKey. There is no more data to be scanned.");
        }

        List<Map<String, String>> items = new ArrayList<>();
        for(Map<String, AttributeValue> av : response.items()) {
            Map<String, String> subMap = new HashMap<>();
            av.entrySet().forEach(e -> {
                subMap.put(e.getKey(), e.getValue().s());
            });
            items.add(subMap);
        }

        try {
            return APIGatewayV2HTTPResponse.builder().withStatusCode(200).withBody(WorkshopUtils.writeValue(items)).build();
        } catch (Exception e) {
            throw new RuntimeException("Unable to list podcasts!", e);
        }
    }

    private Expression getFilterExpressionFromQueryParams(Map<String, String> params) {

        List<Expression> expressions = new ArrayList<>();
        if (params.containsKey("podcast")) {
            throw new UnsupportedOperationException("Podcast condition not implemented");
        }

        if (params.containsKey("in-title")) {
            throw new UnsupportedOperationException("in-title condition not implemented");
        }

        if (expressions.size() == 1) {
            return expressions.get(0);
        } else {
            return Expression.join(expressions.get(0), expressions.get(1), "AND");
        }
    }
}
