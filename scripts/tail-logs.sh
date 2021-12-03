#!/bin/bash
set -e

SDK_STACK_NAME="AwsCrossSdkWorkshop"

HANDLER_NAME=$(aws cloudformation describe-stack-resources \
  --stack-name  "$SDK_STACK_NAME" \
  --query "StackResources[?ResourceType==\`AWS::Lambda::Function\` && contains(LogicalResourceId, \`$1\`)].PhysicalResourceId" \
  --output text
)
aws logs tail "/aws/lambda/$HANDLER_NAME" --format short