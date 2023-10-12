#!/bin/bash

docker compose down

docker compose up -t 0 --detach --build postgresql

sleep 5

docker compose run -v $(pwd)/test.js:/test.js --rm mongo 'mongodb://host.docker.internal:27017/' /test.js

docker compose down

export FERRETDB_POSTGRESQL_NEW=true

docker compose up -t 0 --detach --build postgresql

sleep 5
