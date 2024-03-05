#!/bin/sh

set -ex

pip3 install -r requirements.txt

python3 pymongo_test.py mongodb://localhost:27017/
