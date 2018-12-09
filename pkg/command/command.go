// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"context"

	"github.com/neovim/go-client/nvim"
	"golang.org/x/sync/syncmap"

	"github.com/zchee/nvim-go/pkg/buildctxt"
	"github.com/zchee/nvim-go/pkg/logger"
)

// Command represents a nvim-go plugins commands.
type Command struct {
	Nvim         *nvim.Nvim
	buildContext *buildctxt.Context
	errs         *syncmap.Map
}

// NewCommand return the new Command type with initialize some variables.
func NewCommand(ctx context.Context, v *nvim.Nvim, bctxt *buildctxt.Context) *Command {
	ctx = logger.NewContext(ctx, logger.FromContext(ctx).Named("command"))

	return &Command{
		Nvim:         v,
		buildContext: bctxt,
		errs:         new(syncmap.Map),
	}
}
