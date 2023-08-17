#!/bin/sh

set -ex

npm ci

env CONNECTION_STRING=mongodb://localhost:27017 npx jest -i
