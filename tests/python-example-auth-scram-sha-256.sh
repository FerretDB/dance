#!/bin/sh

set -ex

pip install --user pipenv

. $(pipenv --venv)/bin/activate
pipenv install

python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256'
