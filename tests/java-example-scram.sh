#!/bin/sh

set -ex

export MAVEN_OPTS='-enableassertions'

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
    -Dexec.args="mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-1"

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
    -Dexec.args="mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=SCRAM-SHA-256"
