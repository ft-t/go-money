#!/bin/bash

COMPOSE_DIR="/root/go-money/compose"
SOURCE_DATA="/root/go-money/compose2/db_data"

cd $COMPOSE_DIR
docker compose down
rm -rf db_data
cp -r $SOURCE_DATA db_data
docker compose up -d
