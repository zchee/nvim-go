#!/bin/sh

touch $NVIM_GO_LOG_FILE
exec $1/bin/nvim-go 2>> $NVIM_GO_LOG_FILE
