#!/usr/bin/env bash

set -e

podman run --rm \
  --name cloud3-api \
  -v "$PWD:/app" \
  -w /app \
  -p 8080:8080 \
  docker.io/library/golang:1.25 \
  go "$@"
