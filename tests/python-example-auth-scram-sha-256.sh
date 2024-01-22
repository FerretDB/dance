#!/bin/sh

set -ex

pip3 install pymongo

python3 pymongo_test.py 'mongodb://user:password@localhost:27017/?directConnection=true&tls=true&tlsCertificateKeyFile=../../build/certs/mongodb.pem&tlsCaFile=../../build/certs/mongodb-ca.crt&authMechanism=SCRAM-SHA-256'
