package engine

import (
	"context"
)

const (
	optionsKey = "deploy.options"
)

// Options are runtime configurations that are disseminated to every shipper
// via the context passed to `Shipper.ShipIt` and `Shipper.Rollback`.
type Options struct {
	AppEngine struct {
		ImageTag string
	}

	Version string
}

// InContext embeds the Options into the ctx argument and returns a new
// context.
func (opts *Options) InContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, optionsKey, opts)
}

// ContectForOptions builds a new contexts from context.Background with the
// Options included.
func ContextForOptions(opts Options) context.Context {
	return opts.InContext(context.Background())
}

// OptionsFromContext extracts the Options from a context option, or returns
// an empty Options.
func OptionsFromContext(ctx context.Context) Options {
	if opts, ok := ctx.Value(optionsKey).(*Options); ok {
		return *opts
	}

	return Options{}
}
