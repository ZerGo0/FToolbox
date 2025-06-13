#!/bin/bash

# root check
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

docker load < image.tar
# mv docker-compose.prod.yml docker-compose.yml
#docker compose up -d --remove-orphans
#rm image.tar