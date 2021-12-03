import os
import json

import boto3

_DYNAMODB_CLIENT = None
_S3_CLIENT = None

ENV_PREFIX = 'AWS_SDK_WORKSHOP_'
PODCAST_TABLENAME = os.environ.get(ENV_PREFIX + 'PODCAST_EPISODE_TABLE_NAME')
PODCAST_BUCKET = os.environ.get(ENV_PREFIX + 'PODCAST_DATA_BUCKET_NAME')
S3_PODCAST_PREFIX = os.environ.get(ENV_PREFIX + 'PODCAST_DATA_KEY_PREFIX')


HOUR_IN_SECONDS = 60 * 60


def get_dynamodb_client():
    global _DYNAMODB_CLIENT
    if _DYNAMODB_CLIENT is None:
        _DYNAMODB_CLIENT = boto3.resource('dynamodb').meta.client
    return _DYNAMODB_CLIENT


def get_s3_client():
    global _S3_CLIENT
    if _S3_CLIENT is None:
        _S3_CLIENT = boto3.client('s3')
    return _S3_CLIENT


def get_not_found_error_response(message=''):
    return get_error_response(404, 'NotFoundError', message)


def get_too_many_requests_error_response(message=''):
    return get_error_response(429, 'TooManyRequestsError', message)


def get_error_response(status_code, error_code, message=''):
    return {
        'statusCode': status_code,
        'body': json.dumps(
            {
                'Code': error_code,
                'Message': f'{error_code}: {message}'
            }
        )
    }


def get_redirect_response(location):
    return {
        'statusCode': 307,
        'headers': {
            'location': location
        }
    }