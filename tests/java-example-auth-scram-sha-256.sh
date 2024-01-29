#!/bin/sh

set -ex

# enables Maven exceptions
export MAVEN_OPTS='-ea'

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
-Dexec.args="mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256"
