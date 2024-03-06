#!/bin/sh

set -ex

pip3 install pymongo

python3 pymongo_test.py 'mongodb://username:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256'
