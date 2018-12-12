package scripts

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ScriptShipper runs scripts from a well known location on the file system.
type ScriptShipper struct {
	Opts map[string]interface{}
}

func (ss ScriptShipper) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		defer failSafe("ScriptShipper", ch)

		scriptDir := ss.Opts["scriptDirectory"].(string)
		filenames := ss.Opts["filenames"].([]string)

		for _, filename := range filenames {
			absPath := filepath.Abs(scriptDir + "/" + filename)

			if !strings.HasPrefix(absPath, scriptDir) {
				errMsg := fmt.Sprintf(
					"Can't execute file %v because it lies outside of %v",
					absPath,
					scriptDir,
				)
				return errors.New(errMsg)
			}

			bash := exec.Command("bash")
			bash.Stdout = os.Stdout
			bash.Stderr = os.Stderr
		}
	}()

	return ch
}

func (ss ScriptShipper) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}
