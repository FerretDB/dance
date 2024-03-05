#!/bin/sh

set -ex

pip install --user pipenv

pipenv --python $(python --version) | awk '{ print $NF }'

. $(pipenv --venv)/bin/activate 
pipenv install

python3 pymongo_test.py mongodb://localhost:27017/
