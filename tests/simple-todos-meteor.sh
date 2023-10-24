#!/bin/sh

set -ex

npm install

curl https://install.meteor.com/ | sh

meteor update

export MONGO_URL="mongodb://localhost:27017/"

# directConnection=true avoids server discovery
export MONGO_OPLOG_URL="mongodb://localhost:27017/local?replicaSet=rs0&directConnection=true"

meteor run
