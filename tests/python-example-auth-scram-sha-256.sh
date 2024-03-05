#!/bin/sh

set -ex

pip install --user pipenv

pipenv --python 3.6.4

. $(pipenv --venv)/bin/activate
pipenv install

python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256'
