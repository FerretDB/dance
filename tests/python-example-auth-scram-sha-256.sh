#!/bin/sh

set -ex

pip install --user pipenv

pipenv --python $(python3 -c 'import sys; print(".".join(map(str, sys.version_info[:3])))')

. $(pipenv --venv)/bin/activate
pipenv install

python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256'
