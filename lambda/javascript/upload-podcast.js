import { Upload } from '@aws-sdk/lib-storage';
import { request } from 'https';
import mime from 'mime';
import { getS3Client, PODCAST_BUCKET, S3_PODCAST_PREFIX } from './utils.js';

export const handler = async (event) => {
  console.log(`Event:\n${JSON.stringify(event)}`);

  const bucket = PODCAST_BUCKET;
  const key = getPodcastUploadKey(event);
  const s3Client = getS3Client();

  const mediaUrl = event?.episode?.media_url;
  console.info(`Downloading podcast from: ${mediaUrl}`);
  const incomeMessage = await downloadFromUrl(mediaUrl);
  const contentType = resolveContentType(event, mediaUrl, incomeMessage);
  console.log(
    `Uploading podcast to S3 URI: s3://${bucket}/${key} with Content-Type: ${contentType}`
  );
  throw new Error("Need to implement s3 upload");

  return updateEventState(event, key, contentType);
};

const getPodcastUploadKey = (event) => {
  return `${S3_PODCAST_PREFIX}${event?.episode?.id}/raw-media`;
};

const downloadFromUrl = (url) =>
  new Promise((resolve, reject) => {
    const req = request(url, (res) => {
      resolve(res);
    });
    req.on('error', (err) => {
      reject(err);
    });
    req.end();
  });

const resolveContentType = (event, mediaUrl, response) => {
  let contentType = event?.episode?.media_content_type;
  if (contentType === '' || contentType === 'application/octet-stream') {
    contentType = response.headers['Content-Type'];
  }
  contentType = contentType ?? mime.getType(mediaUrl);

  if (!contentType) {
    throw new Error(`Could not resolve content type for podcast`);
  }
  return contentType;
};

const updateEventState = (event, key, contentType) => {
  return {
    ...event,
    episode: {
      ...event.episode,
      media_key: key,
      media_content_type: contentType,
    },
  };
};
