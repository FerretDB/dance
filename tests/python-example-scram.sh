#!/bin/sh

set -ex

python3 -m venv .

./bin/pip3 install -r requirements.txt

./bin/python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-1'

./bin/python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256'
