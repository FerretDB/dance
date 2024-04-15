#!/bin/sh

set -ex

export MAVEN_OPTS='-enableassertions'

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
    -Dexec.args="mongodb://username:password@localhost:27017/?authMechanism=PLAIN"
