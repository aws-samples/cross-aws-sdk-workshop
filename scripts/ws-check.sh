#!/bin/bash

usage() {
  cat 1>&2 <<EOF
Check your progress in the workshop against the final versions.

USAGE:
    ./ws-check.sh [-d | --diff] [-p | --print] [-c | --copy] REF FILENAME

The --diff option will diff a file in your workspace against the final version.

    ./ws-check.sh --diff origin/final ./scripts/tail-events.sh

The --print option will print the final version of a final.

    ./ws-check.sh --print origin/final ./scripts/tail-events.sh

The --copy option will copy the final version of a file into your current workspace.

    ./ws-check.sh --copy origin/final ./scripts/tail-events.sh

EOF
}

if [ -z "$1" ]
then
  usage
  exit 1
fi

key="$1"
REF="$2"
FILENAME="$3"

case $key in
  -h|--help)
    usage
    exit 0
    ;;
  -d|--diff)
    git diff -R ${REF} -- "${FILENAME}"
    shift
    shift
    ;;
  -p|--print)
    REF="$2"
    FILENAME="$3"
    git show "${REF}:${FILENAME}"
    shift
    shift
    ;;
  -c|--copy)
    REF="$2"
    FILENAME="$3"
    git show "${REF}:${FILENAME}" > "${FILENAME}"
    shift
    shift
    ;;
  *)
    echo "Unknown option $1" 1>&2
    echo
    shift
    usage
    exit 1
    ;;
esac
