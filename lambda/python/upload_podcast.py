import os
import contextlib
import mimetypes
import urllib.parse

import urllib3

from utils import PODCAST_BUCKET, S3_PODCAST_PREFIX, get_s3_client


class UnresolvableContentTypeException(Exception):
    pass


def lambda_handler(event, context):
    s3_bucket = PODCAST_BUCKET
    s3_key = _get_s3_podcast_upload_key(event)
    s3_client = get_s3_client()
    print(f'Downloading podcast from: {event["episode"]["media_url"]}')
    with _downloaded_podcast_stream(event['episode']['media_url']) as stream:
        content_type = _resolve_content_type(event, stream)
        print(
            f'Uploading podcast to S3 URI: s3://{s3_bucket}/{s3_key} '
            f'with Content-Type: {content_type}'
        )
        raise NotImplementedError(
            'Need to call s3_client.upload_fileobj() method to upload to S3'
        )
    _update_event_state(event, s3_key, content_type)
    return event


def _get_s3_podcast_upload_key(event):
    path = urllib.parse.urlparse(event['episode']['media_url']).path
    return f'{S3_PODCAST_PREFIX}{event["episode"]["id"]}/raw-media'


@contextlib.contextmanager
def _downloaded_podcast_stream(podcast_uri):
    http = urllib3.PoolManager()
    resp = http.request(
        "GET", podcast_uri, preload_content=False,
    )
    yield resp
    resp.release_conn()


def _resolve_content_type(event, response):
    content_type = event['episode']['media_content_type']
    if content_type in ['', 'application/octet-stream']:
        content_type = response.headers.get('Content-Type', '')
    if not content_type:
        content_type = mimetypes.guess_type(response.url)
    if not content_type:
        raise UnresolvableContentTypeException(
            'Could not resolve content type for podcast'
        )
    return content_type


def _update_event_state(event, s3_key, content_type):
    event['episode']['media_key'] = s3_key
    event['episode']['media_content_type'] = content_type
