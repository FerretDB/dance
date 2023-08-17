#!/bin/sh

set -ex

pip3 install pymongo

python3 pymongo_test.py mongodb://localhost:27017/
