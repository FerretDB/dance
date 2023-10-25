#!/bin/sh

set -ex

npm install -g npm@latest

export PUPPETEER_DOWNLOAD_PATH=~/.npm/chromium

# TODO add URL environment variables
# TODO add more oplog tests
TINYTEST_FILTER="mongo-livedata - oplog - cursorSupported" ./packages/test-in-console/run.sh
