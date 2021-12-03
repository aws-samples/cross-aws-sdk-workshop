#!/bin/bash
set -e

DEFAULT_NUM_ITEMS=3000
NUM_ITEMS=${1:-$DEFAULT_NUM_ITEMS}
NUM_PUT_ITEM_BATCHES=100
TMP_WRITE_CAPACITY=20
SDK_STACK_NAME="AwsCrossSdkWorkshop"

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

TABLE_NAME=$(aws cloudformation describe-stack-resources \
  --stack-name "$SDK_STACK_NAME" \
  --query 'StackResources[?starts_with(LogicalResourceId, `PodcastEpisode`)].PhysicalResourceId' \
  --output text)

read -r CURRENT_READ_CAPACITY CURRENT_WRITE_CAPACITY < <(aws dynamodb describe-table \
  --table-name "$TABLE_NAME" \
  --query 'Table.ProvisionedThroughput.[ReadCapacityUnits,WriteCapacityUnits]' \
  --output text)

export AWS_MAX_ATTEMPTS=10

generate_and_add_items() {
  START_I="$1"
  END_I="$2"
  DATA="["
  for (( i=START_I; i<END_I; i++ ))
  do
    DATA+=$(cat ${SCRIPT_DIR}/sample-item.json | sed "s|REPLACE|episode-$i|g")
    if [ "$i" -ne "$END_I" ]; then
      DATA+=','
    fi
  done
  DATA+="]"

  echo "Adding items $START_I to $((END_I-1)) to DynamoDB table: $TABLE_NAME"
  echo "$DATA" | aws ddb put "$TABLE_NAME" -
}

echo "Setting capacity from $CURRENT_WRITE_CAPACITY to $TMP_WRITE_CAPACITY for DynamoDB table: $TABLE_NAME"
aws dynamodb update-table \
  --table-name "$TABLE_NAME" \
  --provisioned-throughput ReadCapacityUnits="$CURRENT_READ_CAPACITY",WriteCapacityUnits="$TMP_WRITE_CAPACITY" > /dev/null

echo "Starting process of populating table: $TABLE_NAME with $NUM_ITEMS items"
for (( start=1; start<NUM_ITEMS; start=start+NUM_PUT_ITEM_BATCHES ))
do
  generate_and_add_items "$start" $((start+NUM_PUT_ITEM_BATCHES))
done

echo "Population complete. Setting write capacity back to $CURRENT_WRITE_CAPACITY"
aws dynamodb update-table \
  --table-name "$TABLE_NAME" \
  --provisioned-throughput ReadCapacityUnits="$CURRENT_READ_CAPACITY",WriteCapacityUnits="$CURRENT_WRITE_CAPACITY" > /dev/null