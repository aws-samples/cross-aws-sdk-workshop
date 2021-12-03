#!/bin/bash
set -e

SDK_STACK_NAME="AwsCrossSdkWorkshop"

TABLE_NAME=$(aws cloudformation describe-stack-resources \
  --stack-name "$SDK_STACK_NAME" \
  --query 'StackResources[?starts_with(LogicalResourceId, `PodcastEpisode`)].PhysicalResourceId' \
  --output text)

echo "Resetting data and properties on DynamoDB table: $TABLE_NAME"
aws dynamodb delete-table --table-name $TABLE_NAME > /dev/null
aws dynamodb wait table-not-exists --table-name $TABLE_NAME
aws dynamodb create-table --table-name $TABLE_NAME \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 > /dev/null
aws dynamodb wait table-exists --table-name $TABLE_NAME
echo "Successfully reset DynamoDB table: $TABLE_NAME"
