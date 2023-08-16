#!/bin/sh

set -ex

npm i

env CONNECTION_STRING=mongodb://localhost:27017 npx jest -i
