#!/bin/sh

touch $NVIM_GO_LOG_FILE

if [[ -f $1/bin/nvim-go-race ]]; then
  FLAG=-race
fi
exec $1/bin/nvim-go$FLAG 2>> $NVIM_GO_LOG_FILE
