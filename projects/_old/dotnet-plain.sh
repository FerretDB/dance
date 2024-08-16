#!/bin/sh

set -ex

dotnet run 'mongodb://user:password@localhost:27017/?replicaSet=rs0&authMechanism=PLAIN'
