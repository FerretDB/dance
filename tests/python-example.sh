#!/bin/sh

set -ex

pip install --user pipenv

. $(pipenv --venv)/bin/activate 
pipenv install

python3 pymongo_test.py mongodb://localhost:27017/
