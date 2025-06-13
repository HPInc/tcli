#!/bin/sh
LINTER_IMAGE=golangci/golangci-lint:v2.1.6-alpine
DIR=${1:-$(pwd)}

# flag overrides
docker run --rm \
  -v "$DIR":/tcli:ro \
  -w /tcli \
  "$LINTER_IMAGE" golangci-lint run \
  -v --tests=false --timeout=5m
