import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { DynamoDBDocumentClient } from '@aws-sdk/lib-dynamodb';
import { S3Client } from "@aws-sdk/client-s3";
import { env } from 'process';

const ENV_PREFIX = 'AWS_SDK_WORKSHOP_';
export const PODCAST_TABLENAME = env[`${ENV_PREFIX}PODCAST_EPISODE_TABLE_NAME`];
export const PODCAST_BUCKET = env[`${ENV_PREFIX}PODCAST_DATA_BUCKET_NAME`];
export const S3_PODCAST_PREFIX = env[`${ENV_PREFIX}PODCAST_DATA_KEY_PREFIX`];
export const TRANSCRIBE_STATE_MACHINE_ARN = env[`${ENV_PREFIX}TRANSCRIBE_STATEMACHINE_ARN`];
export const PODCAST_DATA_KEY_PREFIX = env[`${ENV_PREFIX}PODCAST_DATA_KEY_PREFIX`];

let dynamodbClient;
export const getDynamodbClient = () => {
  if (!dynamodbClient) {
    dynamodbClient = new DynamoDBDocumentClient(new DynamoDBClient({}));
  }
  return dynamodbClient;
};

let s3Client;
export const getS3Client = () => {
  if (!s3Client) {
    s3Client = new S3Client({});
  }
  return s3Client;
};

export const badRequestErrorResponse = (message) => ({
  statusCode: 400,
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    Code: 'BadRequestError',
    Message: `BadRequestError: ${message}`,
  }),
});

export const notFoundErrorResponse = (message) => ({
  statusCode: 404,
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    Code: 'NotFoundError',
    Message: `NotFoundError: ${message}`,
  }),
});

export const tooManyRequestsErrorResponse = (message) => ({
  statusCode: 429,
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    Code: 'TooManyRequestsError',
    Message: `TooManyRequestsError: ${message}`,
  }),
});

export const temporaryRedirectResponse = (location) => ({
  statusCode: 307,
  headers: {
    location
  }
});
