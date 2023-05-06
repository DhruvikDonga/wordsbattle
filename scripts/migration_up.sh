#!/bin/bash

echo "Running upward migration"

echo "Starting docker container utility"

#docker-compose run --entrypoint="" migrate ls \migration #check the files in container volume
docker compose up migrate

