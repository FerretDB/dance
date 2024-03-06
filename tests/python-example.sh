#!/bin/sh

set -ex

sudo apt install python3.10-venv

python3 -m venv .

pip3 install -r requirements.txt

python3 pymongo_test.py mongodb://localhost:27017/
