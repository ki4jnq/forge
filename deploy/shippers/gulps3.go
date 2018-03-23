package shippers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	ErrMissingAccessKeys = errors.New("Missing AWS Access Keys")
)

type GulpS3Shipper struct {
	Opts map[string]interface{}
}

func (gss *GulpS3Shipper) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	go func() {
		if err := gss.runGulp(os.Stdout, os.Stderr); err != nil {
			fmt.Println("Failed to run command")
			fmt.Println(err)
			ch <- err
		}
		close(ch)
	}()
	return ch
}

func (gss *GulpS3Shipper) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (gss *GulpS3Shipper) runGulp(stdout, stderr io.Writer) error {
	cmd := exec.Command("gulp", "build")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (gss *GulpS3Shipper) runS3Copy(stdout, stderr io.Writer) error {
	cmd := exec.Command("aws", "s3", "cp", "", "")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func (gss *GulpS3Shipper) KeyPair() (string, string, error) {
	pub, pubOk := gss.Opts["aws_access_key_id"]
	prv, prvOk := gss.Opts["aws_secret_access_key"]
	if !pubOk || !prvOk {
		return "", "", ErrMissingAccessKeys
	}
	// Hmmm... this panics if the pub/prv keys are not strings.
	return pub.(string), prv.(string), nil
}
