#!/bin/sh

set -ex

npm ci

env CONNECTION_STRING=mongodb://username:password@localhost:27017 npx jest -i
