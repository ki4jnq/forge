package options

import (
	"context"
)

const (
	optionsKey = "deploy.options"
)

type Options struct {
	AppEngine struct {
		ImageTag string
	}

	Version string
}

func (opts *Options) InContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, optionsKey, opts)
}

func FromContext(ctx context.Context) Options {
	if opts, ok := ctx.Value(optionsKey).(*Options); ok {
		return *opts
	}

	return Options{}
}
