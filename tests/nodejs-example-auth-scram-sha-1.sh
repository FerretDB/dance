#!/bin/sh

set -ex

npm install

node index.js --uri='mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-1'
