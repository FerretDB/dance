#!/bin/sh

set -ex

dotnet run 'mongodb://localhost:27017/?replicaSet=rs0'
