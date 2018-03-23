package shippers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type ShellShipper struct {
	Opts map[string]interface{}
}

func (shsh *ShellShipper) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	steps := shsh.Opts["steps"].([]interface{})
	go func() {
		defer close(ch)
		defer failSafe("ShellShipper", ch)

		for _, s := range steps {
			// Check in between steps to see if the context has been canceled. If
			// it has, stop processing work.
			select {
			case <-ctx.Done():
				return
			default: // Don't block.
			}

			step, ok := s.(string)
			if !ok {
				ch <- errors.New(fmt.Sprintf("ERROR: Failed to convert step to string: %v\n", step))
				return
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
