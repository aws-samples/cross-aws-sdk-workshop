import { paginateScan, ScanCommand } from '@aws-sdk/lib-dynamodb';
import { unmarshall } from '@aws-sdk/util-dynamodb';
import {
  contains,
  equals,
  ExpressionAttributes,
  serializeConditionExpression,
  serializeProjectionExpression,
} from '@aws/dynamodb-expressions';
import { getDynamodbClient, PODCAST_TABLENAME } from './utils.js';

export const handler = async (event) => {
  console.log(`Event:\n${JSON.stringify(event)}`);

  const client = getDynamodbClient();
  const podcasts = [];
  for await (const record of paginateScan(
    { client },
    getScanParams(event.queryStringParameters)
  )) {
    if (record.Items) {
      podcasts.push(...record.Items);
    }
    logResponseInfo(record);
  }
  return podcasts;
};

const getScanParams = (queryString) => {
  console.log(
    `Received the following query string parameters: ${JSON.stringify(
      queryString
    )}`
  );
  const attributes = new ExpressionAttributes();
  const projectionExpression = serializeProjectionExpression(
    ['id', 'title', 'podcast'],
    attributes
  );
  const conditionalExpression = { type: 'And', conditions: [] };
  if (queryString?.podcast) {
    // TODO: Use the conditional expression from @aws/dynamodb-expressions
    // package to set the condition of the "podcast" attribute equaling the
    // value of "podcast" parameter in the query string.
    throw new Error("Conditional expression for podcast is not implemented");
  }
  if (queryString?.['in-title']) {
    // TODO: Use the conditional expression from @aws/dynamodb-expressions
    // package to set the condition of the "title" attribute equaling the value
    // of the "in-title" parameter in the query string.
    throw new Error("Conditional expression for in-title is not implemented");
  }
  let filterExpression = serializeConditionExpression(
    conditionalExpression, 
    attributes
  );
  if (!filterExpression?.length > 0) {
    filterExpression = undefined;
  }
  // Attribute Values is only need when filter expressions exist.
  const expressionAttributeValues = filterExpression
    ? attributes.values
    : undefined;

  return {
    TableName: PODCAST_TABLENAME,
    FilterExpression: filterExpression,
    ProjectionExpression: projectionExpression,
    ExpressionAttributeNames: attributes.names,
    ExpressionAttributeValues: expressionAttributeValues,
  };
};

const logResponseInfo = (response) => {
  console.log(`Number of podcasts returned in response: ${response.Count}`);
  if (response.LastEvaluatedKey) {
    console.log(
      'Response contains LastEvaluatedKey. There is still more data to be scanned.'
    );
  } else {
    console.log(
      'Response does not contain LastEvaluatedKey. There is no more data to be scanned.'
    );
  }
};
