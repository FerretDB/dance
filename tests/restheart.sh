#!/bin/sh

set -ex

./mvnw clean verify -DskipUTs -P-mongodb \
-Dtest-connection-string="mongodb://user:password@localhost/ferretdb?authMechanism=PLAIN" \
-Dkarate.options="--tags ~@requires-replica-set"
