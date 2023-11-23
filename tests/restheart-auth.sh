#!/bin/sh

set -ex

unset JAVA_HOME

./mvnw clean package

./mvnw clean verify -DskipUTs -P-mongodb -Dtest-connection-string="mongodb://user:password@localhost/ferretdb?authMechanism=PLAIN"
