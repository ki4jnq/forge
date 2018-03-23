package shippers

import "context"

type S3Copy struct{}

func (sc *S3Copy) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (sc *S3Copy) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}
