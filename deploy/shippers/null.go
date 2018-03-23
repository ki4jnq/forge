package shippers

import "context"

// NullShipper is a shipper that Always succeeds and never does a darn thing.
type NullShipper struct{}

func (ns NullShipper) ShipIt(ctx context.Context) chan error {
	ch := make(chan error)
	defer close(ch) // Make a channel and close it so no one blocks on this.
	return ch
}

func (ns NullShipper) Rollback(ctx context.Context) chan error {
	ch := make(chan error)
	close(ch)
	return ch
}
