#!/bin/bash
set -e

if [ ! -z $1 ]; then
  TEMP_DIR=$1$(mktemp -u -d -t mats-XXXXXXXXXX)/
else
  TEMP_DIR=$(mktemp -d -t mats-XXXXXXXXXX)/
fi

REPO_PATHS=$(curl -s https://gitstar-ranking.com/repositories | pup 'a.list-group-item.paginated-item attr{href}')
for REPO_PATH in $REPO_PATHS; do
  REPO_URL=https://github.com${REPO_PATH}
  echo "Cloning" $REPO_PATH "into" $TEMP_DIR
  git clone $REPO_URL $TEMP_DIR
  COUNT=1000 go run cmd/import/main.go $TEMP_DIR $REPO_PATH
  rm -rf $TEMP_DIR
done
