#!/bin/sh

set -ex

export PUPPETEER_DOWNLOAD_PATH=~/.npm/chromium

export METEOR_LOCAL_DIR=~/.meteor

export MONGO_URL="mongodb://localhost:27017/"

export MONGO_OPLOG_URL="mongodb://localhost:27017/local?replicaSet=rs0"

TINYTEST_FILTER="mongo-livedata - oplog - x - implicit collection creation" ./packages/test-in-console/run.sh --once
