import { DynamoDBDocumentClient, GetCommand } from '@aws-sdk/lib-dynamodb';
import {
  badRequestErrorResponse,
  getDynamodbClient,
  notFoundErrorResponse,
  PODCAST_TABLENAME,
  tooManyRequestsErrorResponse,
} from './utils.js';

export const handler = async (event) => {
  console.log(`Event:\n${JSON.stringify(event)}`);

  if (!event?.pathParameters?.id) {
    return badRequestErrorResponse('Episode id not provided');
  }

  const client = getDynamodbClient();
  const docClient = new DynamoDBDocumentClient(client);
  let podcast;
  try {
    podcast = await getPodcastFromTable(docClient, PODCAST_TABLENAME, event);
  } catch (e) {
    if (e?.name === 'ProvisionedThroughputExceededException') {
      console.error(`Received exception: ${e}. Returning a 429 HTTP resposne`);
      return tooManyRequestsErrorResponse('Please slow down request rate');
    }
    throw e;
  }

  if (!podcast) {
    return notFoundErrorResponse(
      `Podcast id ${event?.pathParameters?.id} not found`
    );
  }
  return podcast;
};

const getPodcastFromTable = async (client, tableName, event) => {
  const response = await client.send(
    new GetCommand({
      TableName: tableName,
      Key: event.pathParameters,
      ProjectionExpression: 'id, #n, #d, #p, #s',
      ExpressionAttributeNames: {
        '#n': 'title',
        '#d': 'description',
        '#p': 'podcast',
        '#s': 'status',
      },
    })
  );
  return response?.Item;
};
