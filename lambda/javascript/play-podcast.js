import { GetObjectCommand, waitUntilObjectExists } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import {
  badRequestErrorResponse,
  getS3Client,
  notFoundErrorResponse,
  PODCAST_BUCKET,
  PODCAST_DATA_KEY_PREFIX,
  temporaryRedirectResponse,
} from './utils.js';

const EPISODE_CONTENT_KIND_MEDIA = 'media';
const EPISODE_CONTENT_KIND_TEXT = 'text';
const HOUR = 60 * 60;

export const handler = async (event) => {
  console.log(`Event:\n${JSON.stringify(event)}`);

  const episodeId = event?.pathParameters.id;
  if (episodeId === undefined) {
    return badRequestErrorResponse('Episode id not provided');
  }

  const contentType = parseEpisodeContentKind(
    event?.queryStringParameters?.content
  );
  if (contentType === undefined) {
    return badRequestErrorResponse(
      `Unknown content kind ${event?.queryStringParameters?.content}`
    );
  }

  const mediaKey = getEpisodeMediaKey(contentType, episodeId);
  if (mediaKey === undefined) {
    return badRequestErrorResponse(`unknonw content type ${contentType}`);
  }

  const s3Client = getS3Client();
  const mediaExists = await checkMediaExists(
    s3Client,
    PODCAST_BUCKET,
    mediaKey
  );
  if (!mediaExists) {
    return notFoundErrorResponse(
      `Episode ${episodeId} ${contentType} data not found`
    );
  }

  return await responseRedirect(s3Client, PODCAST_BUCKET, mediaKey);
};

const parseEpisodeContentKind = (content) => {
  if ([EPISODE_CONTENT_KIND_MEDIA, '', content].includes(content)) {
    return EPISODE_CONTENT_KIND_MEDIA;
  } else if (content === 'text') {
    return EPISODE_CONTENT_KIND_TEXT;
  }
};

const getEpisodeMediaKey = (contentType, episodeId) => {
  if (contentType === EPISODE_CONTENT_KIND_MEDIA) {
    return `${PODCAST_DATA_KEY_PREFIX}${episodeId}/raw-media`;
  } else if (contentType === EPISODE_CONTENT_KIND_TEXT) {
    return `${PODCAST_DATA_KEY_PREFIX}${episodeId}transcription.txt`;
  }
};

const responseRedirect = async (client, bucket, key) => {
  // TODO: use the @aws-sdk/s3-request-presigner package's getSignedUrl() to
  // create a presigned URL for the GetObject command.
  return temporaryRedirectResponse(
    `https://s3.${await client.config.region()}.amazonaws.com/${bucket}/${key}`
  );
};

const checkMediaExists = async (client, bucket, key) => {
  // TODO: use S3 client(@aws-sdk/client-s3)'s waitUntilObjectExists() to detect
  // if the mediaKey exists, and also allow the lambda handler to wait a short
  // period of time for the object to be created.
  return true;
};
