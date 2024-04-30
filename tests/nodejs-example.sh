#!/bin/sh

set -ex

npm install

node index.js --uri='mongodb://localhost:27017/'
