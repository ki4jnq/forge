package engine

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

func ContextForOptions(opts Options) context.Context {
	return opts.InContext(context.Background())
}

func OptionsFromContext(ctx context.Context) Options {
	if opts, ok := ctx.Value(optionsKey).(*Options); ok {
		return *opts
	}

	return Options{}
}
