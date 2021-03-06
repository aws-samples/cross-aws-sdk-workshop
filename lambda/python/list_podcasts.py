from utils import (
    PODCAST_TABLENAME, get_dynamodb_client
)

import boto3.dynamodb.conditions


def lambda_handler(event, context):
    dynamodb_client = get_dynamodb_client()
    response = dynamodb_client.scan(**_get_scan_kwargs(event))
    _print_response_debugging_information(response)
    return response['Items']


def _print_response_debugging_information(response):
    print(f'Number of podcasts returned in response: {response["Count"]}')
    if 'LastEvaluatedKey' in response:
        print('Response contains LastEvaluatedKey. There is still more data to be scanned.')
    else:
        print('Response does not contain LastEvaluatedKey. There is no more data to be scanned.')


def _get_scan_kwargs(event):
    scan_kwargs = {
        'TableName': PODCAST_TABLENAME,
        'ProjectionExpression': 'id, #t, #p',
        'ExpressionAttributeNames': {
            '#t': 'title',
            '#p': 'podcast',
        },
    }
    if event.get('queryStringParameters'):
        scan_kwargs['FilterExpression'] = _get_filter_expression_from_querystring(
            event['queryStringParameters']
        )
    return scan_kwargs


def _get_filter_expression_from_querystring(qs_params):
    print(f'Received the following query string parameters: {qs_params}')
    podcast_condition = None
    in_title_condition = None
    if 'podcast' in qs_params:
        # Using the ``boto3.dynamodb.condition.Attr`` class, set
        # ``podcast_condition`` to the condition of the 'podcast' attribute
        # equaling the value of 'podcast' parameter in the query string
        raise NotImplementedError('podcast_condition')
    if 'in-title' in qs_params:
        # Using the ``boto3.dynamodb.condition.Attr`` class,  set
        # ``in_title_condition`` to the condition of the 'title' attribute
        # containing the value of 'in-title' parameter in the query string
        raise NotImplementedError('in_title_condition')

    return _chain_conditions(podcast_condition, in_title_condition)


def _chain_conditions(*conditions):
    current_condition = None
    for condition in conditions:
        if condition:
            if current_condition is None:
                current_condition = condition
            else:
                current_condition = current_condition & condition
    return current_condition
