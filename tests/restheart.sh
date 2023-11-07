#!/bin/sh

set -ex

unset JAVA_HOME

./mvnw clean package

./mvnw -X clean verify -DskipUTs -P-mongodb \
-Dtest-connection-string="mongodb://localhost/ferretdb" \
-Dkarate.options="--tags ~@requires-replica-set"
