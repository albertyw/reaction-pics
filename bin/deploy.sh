#!/bin/bash

# This script will build and deploy a new docker image

set -exuo pipefail
IFS=$'\n\t'

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd "$DIR"/.. || exit 1

CONTAINER="reaction-pics"
PORT="5003"
NETWORK="$CONTAINER"_net
DEPLOY_BRANCH="${1:-}"
set +x  # Do not print contents of .env
source .env
set -x

if [ -n "$DEPLOY_BRANCH" ]; then
    # Update repository
    git fetch -tp
    git checkout "$DEPLOY_BRANCH"
    git pull
fi
BRANCH="$(git rev-parse --abbrev-ref HEAD)"

# Build and network
docker build --pull -t "$CONTAINER:$BRANCH" .
docker network inspect "$NETWORK" &>/dev/null ||
    docker network create --driver bridge "$NETWORK"

# Start container
docker stop "$CONTAINER" || true
docker container rm "$CONTAINER" || true
docker run \
    --detach \
    --restart=always \
    --publish="127.0.0.1:$PORT:$PORT" \
    --network="$NETWORK" \
    --mount type=bind,source="$(pwd)"/logs/app,target=/root/logs/app \
    --name="$CONTAINER" "$CONTAINER:$BRANCH"

if [ "$ENVIRONMENT" = "production" ] && [ "$BRANCH" = "master" ]; then
    # Cleanup docker
    docker system prune --force --filter "until=168h"
    docker volume prune --force

    # Update nginx
    sudo cp config/nginx/app /etc/nginx/sites-enabled/reaction.pics-app
    docker exec nginx /etc/init.d/nginx reload
fi
