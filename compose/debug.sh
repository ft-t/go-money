#!/bin/bash
docker compose down
cd .. && make build-docker
cd compose && docker compose -f docker-compose-backend.yaml up
