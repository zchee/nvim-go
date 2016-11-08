#!/bin/sh

touch $NVIM_GO_LOG_FILE
exec ./bin/nvim-go "$@" 2>> $NVIM_GO_LOG_FILE
