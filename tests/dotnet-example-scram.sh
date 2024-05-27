#!/bin/sh

set -ex

dotnet run 'mongodb://username:password@localhost:27017/?authMechanism=SCRAM-SHA-1'

dotnet run 'mongodb://username:password@localhost:27017/?authMechanism=SCRAM-SHA-256'
