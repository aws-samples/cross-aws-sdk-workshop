from utils import (
    PODCAST_TABLENAME, get_dynamodb_client, get_not_found_error_response,
    get_too_many_requests_error_response
)


def lambda_handler(event, context):
    dynamodb_client = get_dynamodb_client()
    podcast = _get_podcast_from_table(dynamodb_client, PODCAST_TABLENAME, event)
    if podcast is None:
        return get_not_found_error_response('Podcast not found')
    return podcast


def _get_podcast_from_table(dynamodb_client, table_name, event):
    response = dynamodb_client.get_item(
        TableName=table_name,
        Key=_get_table_key_from_event(event),
        ProjectionExpression='id, #n, #d, #p, #s',
        ExpressionAttributeNames={
            '#n': 'title',
            '#d': 'description',
            '#p': 'podcast',
            '#s': 'status',
        }
    )
    return response.get('Item')


def _get_table_key_from_event(event):
    return event['pathParameters']

