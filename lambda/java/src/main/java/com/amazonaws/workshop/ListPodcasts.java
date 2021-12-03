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

        ScanEnhancedRequest.Builder scanEnhancedRequest =
                ScanEnhancedRequest
                        .builder()
                        .addAttributeToProject("title")
                        .addAttributeToProject("id")
                        .addAttributeToProject("podcast");

        if (event.getQueryStringParameters() != null) {
            scanEnhancedRequest.filterExpression(getFilterExpressionFromQueryParams(event.getQueryStringParameters()));
        }

        PageIterable<PodcastEpisode> podcastEpisodes = podcastTable.scan(scanEnhancedRequest.build());

        List<PodcastEpisode> episodes = podcastEpisodes.stream().flatMap(p -> {
            System.out.println(String.format("Number of podcasts returned in response: %d", p.items().size()));

            if (p.lastEvaluatedKey() != null) {
                System.out.println("Response contains LastEvaluatedKey. There is still more data to be scanned.");
            } else {
                System.out.println("Response does not contain LastEvaluatedKey. There is no more data to be scanned.");
            }

            return p.items().stream();
        }).collect(Collectors.toList());

        try {
            return APIGatewayV2HTTPResponse.builder().withStatusCode(200).withBody(WorkshopUtils.writeValue(episodes)).build();
        } catch (Exception e) {
            throw new RuntimeException("Unable to list podcasts!", e);
        }
    }

    private Expression getFilterExpressionFromQueryParams(Map<String, String> params) {

        List<Expression> expressions = new ArrayList<>();
        if (params.containsKey("podcast")) {
            expressions.add(Expression.builder()
                    .expression("#p = :b")
                    .putExpressionName("#p", "podcast")
                    .putExpressionValue(":b", AttributeValue.builder().s(params.get("podcast")).build())
                    .build());
        }

        if (params.containsKey("in-title")) {
            expressions.add(Expression.builder()
                    .expression("contains(#i, :t)")
                    .putExpressionName("#i", "title")
                    .putExpressionValue(":t", AttributeValue.builder().s(params.get("in-title")).build())
                    .build());
        }

        if (expressions.size() == 1) {
            return expressions.get(0);
        } else {
            return Expression.join(expressions.get(0), expressions.get(1), "AND");
        }
    }
}
