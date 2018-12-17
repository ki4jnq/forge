package engine

import "context"

// A Shipper is anything that can ShipIt! How great is that!
type Shipper interface {
	ShipIt(context.Context) chan error
	Rollback(context.Context) chan error
}

type Shippers map[string]Shipper
