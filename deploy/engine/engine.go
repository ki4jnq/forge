package engine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// Engine manages the details of orchestrating deployments across a range of
// deployment targets.
type Engine struct {
	Shippers map[string]Shipper
}

// NewEngine creates a new Engine that manages the provided shippers.
func NewEngine(shippers map[string]Shipper) *Engine {
	return &Engine{
		Shippers: shippers,
	}
}

// Run performs a deployment based on the configured shippers and the provided
// options. A failure in a single shipper will result in a rollback being
// issued across all shippers as well.
func (eng *Engine) Run(opts Options) error {
	ctx := ContextForOptions(opts)

	// Run the deploy and return if everything works.
	finalErr := eng.runDeploy(ctx)
	if finalErr == nil {
		return nil
	}

	fmt.Println(strings.Repeat("*", 80))
	fmt.Println("An error was encountered while deploying the application")
	fmt.Printf("The error message was: %v\n", finalErr)
	fmt.Println(strings.Repeat("*", 80))

	if err := eng.runRollback(ctx); err != nil {
		return nil
	}

	return finalErr
}

// runDeploy runs every shipper's ShipIt method with a shared context and
// returns the error
func (eng *Engine) runDeploy(baseCtx context.Context) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	deployCh := eng.fanIn(func(shipper Shipper) chan error {
		return shipper.ShipIt(ctx)
	})

	for err = range deployCh {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}

	return
}

func (eng *Engine) runRollback(baseCtx context.Context) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	rollbackCh := eng.fanIn(func(shipper Shipper) chan error {
		return shipper.Rollback(ctx)
	})

	for err = range rollbackCh {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}

	return
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
