#!/bin/sh

set -ex

npm install -g npm@latest

export PUPPETEER_DOWNLOAD_PATH=~/.npm/chromium

export METEOR_LOCAL_DIR=~/.meteor

export MONGO_URL="mongodb://localhost:27017/"

# https://github.com/FerretDB/FerretDB/blob/main/cmd/envtool/envtool.go#L155
export MONGO_OPLOG_URL="mongodb://localhost:27017/local?replicaSet=mongodb-rs&directConnection=true"

TINYTEST_FILTER="mongo-livedata - oplog - x - implicit collection creation" ./packages/test-in-console/run.sh --once
