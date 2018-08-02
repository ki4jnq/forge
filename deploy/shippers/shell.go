package shippers

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

type ShellShipper struct {
	Opts map[string]interface{}
}

func (shsh *ShellShipper) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	steps, ok := shsh.Opts["steps"].([]string)
	if !ok {
		errors.New("No steps defined for shellShipper, skipping")
	}

	go func() {
		defer close(ch)
		defer failSafe("ShellShipper", ch)

		for _, step := range steps {
			// Check in between steps to see if the context has been canceled. If
			// it has, stop processing work.
			select {
			case <-ctx.Done():
				return
			default: // Don't block.
			}

			bash := exec.Command("bash", "-c", step)
			bash.Stdout = os.Stdout
			bash.Stderr = os.Stderr

			if err := bash.Run(); err != nil {
				ch <- err
				return
			}
		}
	}()
	return ch
}

func (shsh *ShellShipper) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}
