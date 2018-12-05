#!/bin/sh

if [ -z "$NVIM_GO_LOG_FILE" ]; then
  NVIM_GO_LOG_FILE=/tmp/nvim-go.log
fi
touch "$NVIM_GO_LOG_FILE"

if [ -n "$NVIM_GO_RACE" ] && [ -f "$1/bin/nvim-go-race" ]; then
  exec "$1/bin/nvim-go-race" 2>> "$NVIM_GO_LOG_FILE"
else
  exec "$1/bin/nvim-go" 2>> "$NVIM_GO_LOG_FILE"
fi
