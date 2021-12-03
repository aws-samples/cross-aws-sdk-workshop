import botocore.exceptions

from utils import (
    PODCAST_BUCKET, S3_PODCAST_PREFIX, HOUR_IN_SECONDS, get_s3_client,
    get_redirect_response, get_not_found_error_response
)


def lambda_handler(event, context):
    s3_bucket = PODCAST_BUCKET
    s3_key = _get_s3_artifact_key(event)
    s3_client = get_s3_client()
    artifact_exists = _wait_for_artifact_to_exist(s3_client, s3_bucket, s3_key)
    if not artifact_exists:
        return get_not_found_error_response(
            f'Episode media data not found, {s3_key}')
    url = _get_artifact_url(
        s3_client=s3_client, s3_bucket=s3_bucket, s3_key=s3_key)
    return get_redirect_response(url)


def _get_s3_artifact_key(event):
    artifact_name = 'raw-media'
    if event.get('queryStringParameters', {}).get('content') == 'text':
        artifact_name = 'transcription.txt'
    episode_id = event['pathParameters']['id']
    return f'{S3_PODCAST_PREFIX}{episode_id}/{artifact_name}'


def _get_artifact_url(s3_client, s3_bucket, s3_key):
    # TODO: Return presigned url for the 'get_object' operation using the
    #  s3_client.generate_presigned_url() method. This URL should be valid
    #  for 24 hours.
    return (
        f'https://s3.{s3_client.meta.region_name}.amazonaws.com/'
        f'{s3_bucket}/{s3_key}'
    )


def _wait_for_artifact_to_exist(s3_client, s3_bucket, s3_key):
    # TODO: Create an 'object_exists' waiter and wait at increments
    #  of three seconds for a maximum of six attempts. If the waiter
    #  fails/times out, the function should catch the WaiterError and return
    #  False.
    return True
