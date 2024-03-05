#!/bin/sh

set -ex

. $(pipenv --venv)/bin/activate
pipenv install

python3 pymongo_test.py mongodb://localhost:27017/
