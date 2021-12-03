#!/bin/bash
set -e

WS_CONFIG_FILE=.workshop-config.json

AWS_ACCOUNT=$(aws sts get-caller-identity \
	--query 'Account' \
	--output text)
echo "Using AWS account: ${AWS_ACCOUNT}"

AWS_REGION=$(aws configure get region)
echo "Using AWS region: ${AWS_REGION}"


while [[ true ]]
do
	echo ""
	echo -n "What language would you like to use? [go/python/java/javascript]: "
	read WS_LANGUAGE
	WS_LANGUAGE=$(echo "$WS_LANGUAGE" | awk '{print tolower($0)}')
	
	case ${WS_LANGUAGE} in
		"go")
			;;
		"python")
			;;
		"java")
			;;
		"javascript")
			;;
		*)
			echo "'${WS_LANGUAGE}', is not valid. Select go, python, java, or javascript"
			continue
			;;
	esac
	break
done

echo "Using workshop language: ${WS_LANGUAGE}"

echo "{\"account\": \"${AWS_ACCOUNT}\", \"region\": \"${AWS_REGION}\", \"language\": \"${WS_LANGUAGE}\"}" > ${WS_CONFIG_FILE}

echo "Workshop configuration saved to: ${WS_CONFIG_FILE}"
