#!/bin/sh

set -ex

. $(pipenv --venv)/bin/activate
pipenv install

python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?authMechanism=PLAIN'
