#!/bin/sh

set -ex

pip install --user pipenv

pipenv --python 3.6.4

. $(pipenv --venv)/bin/activate 
pipenv install

python3 pymongo_test.py mongodb://localhost:27017/
