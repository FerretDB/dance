#!/bin/sh

set -ex

npm install -g npm@latest

export PUPPETEER_DOWNLOAD_PATH=~/.npm/chromium

export MONGO_URL="mongodb://localhost:27017/"

export MONGO_OPLOG_URL="mongodb://localhost:27017/local?replicaSet=rs0&directConnection=true"

TINYTEST_FILTER="mongo-livedata - oplog - cursorSupported" ./packages/test-in-console/run.sh --once
TINYTEST_FILTER="mongo-livedata - oplog - entry skipping" ./packages/test-in-console/run.sh --once
TINYTEST_FILTER="mongo-livedata - oplog - x - implicit collection creation" ./packages/test-in-console/run.sh --once
