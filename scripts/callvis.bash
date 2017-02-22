#!/bin/bash
set -e

if [[ -z "$1" ]]; then
  echo "USAGE: $0 [focus pkg (autocmd|commands|config|context|nvimutil|pathutil)]"
  exit 1
fi

if hash go-callvis 2&/dev/null; then
  echo 'no such command: go-callvis'
  echo 'Please go get -u -v github.com/TrueFurby/go-callvis'
  exit 1
fi
if hash dot 2&/dev/null; then
  echo 'no such command: dot'
  echo 'Please install graphviz'
  exit 1
fi

ROOT=$(git rev-parse --show-toplevel)
GOPATH="${ROOT}:${ROOT}/vendor" go-callvis -group pkg -focus $1 cmd/nvim-go | dot -Tsvg -o ${ROOT}/nvim-go.svg

if hash xdg-open 2&>/dev/null; then
  OPEN_CMD=xdg-open
else
  OPEN_CMD=open
fi

$OPEN_CMD ${ROOT}/nvim-go.svg
