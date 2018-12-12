package engine

import (
	"context"
)

type Execution struct {
	ctx context.Context
	cancel context.CancelFunc
}

func NewExecution() {
}

func (exe *Execution) ExecWithContext(f(ctx) chan error) chan error{
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	return 

	rollbackCh := fanIn(eng.Shippers, func(shipper Shipper) chan error {
	})

	// NOTE: All errors will be printed, but only the last error is returned.
	for err = range rollbackCh {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}

	return
}
