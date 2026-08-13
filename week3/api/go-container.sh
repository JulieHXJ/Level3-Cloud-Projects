#!/usr/bin/env bash

set -e

podman run --rm -it \
  --env JWT_SECRET \
  # --env ADMIN_PASSWORD_HASH \
  # --env VIEWER_PASSWORD_HASH \
  --name cloud3-api \
  -v "$PWD:/app" \
  -w /app \
  -p 8080:8080 \
  docker.io/library/golang:1.25 \
  go "$@"
