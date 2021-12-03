set -e
API_URL=$(aws cloudformation describe-stacks \
  --stack-name AwsCrossSdkWorkshop \
	--query 'Stacks[].Outputs[?OutputKey==`APIUrl`][].OutputValue' \
	--output text)
echo "$API_URL"