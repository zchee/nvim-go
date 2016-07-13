package autocmd

import (
	"sync"

	"nvim-go/context"
	"nvim-go/nvim/quickfix"
)

// Autocmd represents a autocmd context.
type Autocmd struct {
	ctxt *context.Context

	qf []*quickfix.ErrorlistData

	bufWritePostChan chan error
	bufWritePreChan  chan error
	wg               sync.WaitGroup

	errors []error
}
