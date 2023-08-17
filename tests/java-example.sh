#!/bin/sh

set -ex

# enables Maven exceptions
export MAVEN_OPTS='-ea'

mvn compile exec:java -Dexec.mainClass=com.start.Connection \
-Dexec.args="mongodb://localhost:27017/"
