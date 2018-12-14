package engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Engine struct {
	Shippers map[string]Shipper

	opts *Options
}

func NewEngine(opts *Options, shippers map[string]Shipper) *Engine {
	return &Engine{
		opts:     opts,
		Shippers: shippers,
	}
}

func (eng *Engine) Run() error {
	var deployErr error
	//var mustRollback bool

	// Run the deploy and return if everything works.
	if err := eng.runDeploy(); err == nil {
		return nil
	}

	fmt.Println(strings.Repeat("*", 80))
	fmt.Println("An error was encountered while deploying the application")
	fmt.Printf("The error message was: %v\n", deployErr)
	fmt.Println(strings.Repeat("*", 80))

	if err := eng.runRollback(); err != nil {
		return nil
	}

	return deployErr
}

// runDeploy runs every shipper's ShipIt method with a shared context and
// returns the error
func (eng *Engine) runDeploy() (err error) {
	ctx, cancel := eng.buildContext()
	defer cancel()

	deployCh := eng.fanIn(func(shipper Shipper) chan error {
		return shipper.ShipIt(ctx)
	})

	// NOTE: All errors will be printed, but only the last error is returned.
	for err = range deployCh {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}

	return
}

func (eng *Engine) runRollback() (err error) {
	ctx, cancel := eng.buildContext()
	defer cancel()

	rollbackCh := eng.fanIn(func(shipper Shipper) chan error {
		return shipper.Rollback(ctx)
	})

	// NOTE: All errors will be printed, but only the last error is returned.
	for err = range rollbackCh {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}

	return
}

func (eng *Engine) buildContext() (context.Context, context.CancelFunc) {
	ctx := eng.opts.InContext(context.Background())
	return context.WithCancel(ctx)
}

// fanIn runs fn against every Shipper and fans in all errors from their
// returned channels onto a single aggregate channel, which it returns.
func (eng *Engine) fanIn(fn func(shipper Shipper) chan error) chan error {
	var wg sync.WaitGroup
	aggregator := make(chan error)

	for target, shipper := range eng.Shippers {
		wg.Add(1)
		go func(target string, shipper Shipper) {
			defer wg.Done()

			fmt.Printf("%v: Running target\n", target)
			for err := range fn(shipper) {
				aggregator <- err
			}
			fmt.Printf("%v: Completed target\n", target)
		}(target, shipper)
	}

	// Wait for all sub processes to finish and send a signal to the parent
	// when they do.
	go func() {
		wg.Wait()
		close(aggregator)
	}()

	return aggregator
}
