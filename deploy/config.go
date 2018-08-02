package deploy

import (
	"context"
	"errors"

	"github.com/ki4jnq/forge/deploy/shippers"
	"github.com/ki4jnq/forge/deploy/shippers/k8"
	"github.com/ki4jnq/rconf"
)

var ErrNotAShipper = errors.New("Undefined shipper requested in Forgefile")

type Ctx context.Context

type Config map[string]shipperBlock

type Options map[string]interface{}

type shipperBlock struct {
	ShipperName string `yaml:"shipper"`
	Opts        Options
}

func rubyConf(b rconf.Binder) {
	b.BlockWithArg("target", func(b rconf.Binder) {
		targetName := b.StringArgAt(0)

		b.BlockWithArg("shipper", func(b rconf.Binder) {
			shipperName := b.StringArgAt(0)

			shipperBlock := shipperBlock{
				ShipperName: shipperName,
				Opts:        make(map[string]interface{}),
			}
			shipperBlock.defineOpts(b)

			conf[targetName] = shipperBlock
		})
	})
}

func (sb *shipperBlock) defineOpts(b rconf.Binder) {
	switch sb.ShipperName {
	case "k8":
	case "shell":
		b.BindStringFn("step", func(step string) {
			// It's OK if this type assertion fails
			steps, _ := sb.Opts["steps"].([]string)
			sb.Opts["steps"] = append(steps, step)
		})
	default:
		panic(ErrNotAShipper)
	}
}

func (sb *shipperBlock) toShipper() Shipper {
	switch sb.ShipperName {
	case "k8":
		return k8.NewK8Shipper(sb.Opts)
	case "shell":
		return &shippers.ShellShipper{Opts: sb.Opts}
	default:
		panic(ErrNotAShipper)
	}
}
