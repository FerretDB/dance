#!/bin/sh

set -ex

# enables Maven exceptions
export MAVEN_OPTS='-ea'

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
-Dexec.args="mongodb://user:password@localhost:27017/?directConnection=true&tls=true&tlsCertificateKeyFile=../../build/certs/mongodb.pem&tlsCaFile=../../build/certs/mongodb-ca.crt&authMechanism=SCRAM-SHA-1"
