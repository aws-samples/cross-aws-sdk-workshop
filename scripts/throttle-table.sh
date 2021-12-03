#!/bin/bash
set -e

SDK_STACK_NAME="AwsCrossSdkWorkshop"
TMP_READ_CAPACITY=1

TABLE_NAME=$(aws cloudformation describe-stack-resources \
  --stack-name "$SDK_STACK_NAME" \
  --query 'StackResources[?starts_with(LogicalResourceId, `PodcastEpisode`)].PhysicalResourceId' \
  --output text)

read -r CURRENT_READ_CAPACITY CURRENT_WRITE_CAPACITY < <(aws dynamodb describe-table \
  --table-name "$TABLE_NAME" \
  --query 'Table.ProvisionedThroughput.[ReadCapacityUnits,WriteCapacityUnits]' \
  --output text)

echo "Setting capacity from $CURRENT_READ_CAPACITY to $TMP_READ_CAPACITY for DynamoDB table: $TABLE_NAME"
aws dynamodb update-table \
  --table-name "$TABLE_NAME" \
  --provisioned-throughput ReadCapacityUnits="$TMP_READ_CAPACITY",WriteCapacityUnits="$CURRENT_WRITE_CAPACITY" > /dev/null

echo "Continually scanning to throttle DynamoDB table: $TABLE_NAME"
aws ddb select "$TABLE_NAME" > /dev/null
scan_rc=$?

function scan_until_interrupt() {
  while [ "$scan_rc" -eq 0 ] || [ "$scan_rc" -eq 254 ]
  do
    aws ddb select "$TABLE_NAME" > /dev/null
    scan_rc="$?"
  done
}
scan_until_interrupt || :

echo "Stopped scan on DynamoDB table. Setting read capacity back to $CURRENT_READ_CAPACITY"
aws dynamodb update-table \
  --table-name "$TABLE_NAME" \
  --provisioned-throughput ReadCapacityUnits="$CURRENT_READ_CAPACITY",WriteCapacityUnits="$CURRENT_WRITE_CAPACITY" > /dev/null