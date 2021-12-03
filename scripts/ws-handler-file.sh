#!/bin/bash
set -e

HANDLER_NAME=$1

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
WS_LANG=$(cat "${SCRIPT_DIR}/../.workshop-config.json" | sed -n 's|.*"language": "\([^"]*\)".*|\1|p')

case $WS_LANG in
	go)
		case $HANDLER_NAME in
			GetPodcast)
				echo "lambda/go/get-podcast/main.go"
				;;
			ListPodcasts)
				echo "lambda/go/list-podcasts/main.go"
				;;
			PlayPodcast)
				echo "lambda/go/play-podcast/main.go"
				;;
			UploadPodcast)
				echo "lambda/go/upload-podcast/main.go"
				;;
			*)
				echo "unknown handler name $HANDLER_NAME"
				exit 1
				;;
		esac
		;;

	python)
		case $HANDLER_NAME in
			GetPodcast)
				echo "lambda/python/get_podcast.py"
				;;
			ListPodcasts)
				echo "lambda/python/list_podcasts.py"
				;;
			PlayPodcast)
				echo "lambda/python/play_podcast.py"
				;;
			UploadPodcast)
				echo "lambda/python/upload_podcast.py"
				;;
			*)
				echo "unknown handler name $HANDLER_NAME"
				exit 1
				;;
		esac
		;;

	java)
		case $HANDLER_NAME in
			GetPodcast)
				echo "lambda/java/src/main/java/com/amazonaws/workshop/GetPodcast.java"
				;;
			ListPodcasts)
				echo "lambda/java/src/main/java/com/amazonaws/workshop/ListPodcasts.java"
				;;
			PlayPodcast)
				echo "lambda/java/src/main/java/com/amazonaws/workshop/PlayPodcast.java"
				;;
			UploadPodcast)
				echo "lambda/java/src/main/java/com/amazonaws/workshop/UploadPodcast.java"
				;;
			*)
				echo "unknown handler name $HANDLER_NAME"
				exit 1
				;;
		esac
		;;

	javascript)
		case $HANDLER_NAME in
			GetPodcast)
				echo "lambda/javascript/get-podcast.js"
				;;
			ListPodcasts)
				echo "lambda/javascript/list-podcasts.js"
				;;
			PlayPodcast)
				echo "lambda/javascript/play-podcast.js"
				;;
			UploadPodcast)
				echo "lambda/javascript/upload-podcast.js"
				;;
			*)
				echo "unknown handler name $HANDLER_NAME"
				exit 1
				;;
		esac
		;;

	*)
		echo "unknown language $WS_LANG"
		exit 1
		;;
esac
