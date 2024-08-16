#!/bin/sh

set -ex

export MAVEN_OPTS='-enableassertions'

mvn compile exec:java -Dexec.mainClass=com.start.Connection -Dexec.args="mongodb://localhost:27017/?replicaSet=rs0"
