#!/bin/sh

set -ex

dotnet run 'mongodb://username:password@localhost:27017/?authMechanism=PLAIN'
