#!/bin/sh

set -ex

# sudo apt install python3.10-venv

python -m venv .

./bin/pip3 install -r requirements.txt

python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-1'
