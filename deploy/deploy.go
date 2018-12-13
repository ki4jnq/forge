package deploy

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ki4jnq/forge"
	"github.com/ki4jnq/forge/deploy/options"
)

var (
	conf  = Config{}
	flags = flag.NewFlagSet("deploy", flag.ExitOnError)
	opts  = &options.Options{}
)

func init() {
	flags.StringVar(
		&opts.AppEngine.ImageTag,
		"ae-image-tag",
		"",
		"The AppEngine Docker image tag to deploy",
	)
	flags.StringVar(
		&opts.Version,
		"version",
		"",
		"The version number to deploy.",
	)

	forge.Register(&forge.Cmd{
		Name:      "deploy",
		Flags:     flags,
		SubConf:   conf,
		SubRunner: run,
	})
}

func run() error {
	var deployErr error
	var mustRollback bool
	ctx, cancel := buildContext()
	defer cancel()

	shippers := make(map[string]Shipper, len(conf))
	for target, block := range conf {
		shippers[target] = block.toShipper()
	}

	errOnce := sync.Once{}
	deployCh := fanIn(shippers, func(shipper Shipper) chan error {
		return shipper.ShipIt(ctx)
	})

	for err := range deployCh {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		errOnce.Do(func() {
			cancel()
			deployErr = err
			mustRollback = true
		})
	}

	// If everything succeeded, then we're done and can safely exit.
	if !mustRollback {
		return nil
	}

	fmt.Println(strings.Repeat("*", 80))
	fmt.Println("An error was encountered while deploying the application")
	fmt.Printf("The error message was: %v\n", deployErr)
	fmt.Println(strings.Repeat("*", 80))

	ctx, cancel = buildContext()
	defer cancel()
	rollbackCh := fanIn(shippers, func(shipper Shipper) chan error {
		return shipper.Rollback(ctx)
	})

	for err := range rollbackCh {
		fmt.Println(err)
	}

	return deployErr
}

// fanIn runs fn against every Shipper and fans in all errors from their
// returned channels onto a single aggregate channel, which it returns.
func fanIn(shippers map[string]Shipper, fn func(shipper Shipper) chan error) chan error {
	var wg sync.WaitGroup
	aggregator := make(chan error)

	for target, shipper := range shippers {
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

func buildContext() (context.Context, context.CancelFunc) {
	ctx := opts.InContext(context.Background())
	return context.WithCancel(ctx)
}
