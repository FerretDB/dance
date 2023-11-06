#!/bin/sh

set -ex

# ensure we set JAVA_HOME to use Java 17
export JAVA_HOME=$(readlink -f /usr/bin/javac | sed "s:/bin/javac::")

JAVA_HOME=$(readlink -f /usr/bin/javac | sed "s:/bin/javac::") ./mvnw -X clean verify -DskipUTs -P-mongodb \
-Dtest-connection-string="mongodb://user:password@localhost/ferretdb?authMechanism=PLAIN" \
-Dkarate.options="--tags ~@requires-replica-set"
