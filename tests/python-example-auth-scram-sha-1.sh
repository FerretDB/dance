#!/bin/sh

set -ex

pip3 install pymongo

python3 pymongo_test.py mongodb://user:password@127.0.0.1:27017/?authMechanism=SCRAM-SHA-1
