package deploy

import (
	"context"
	"errors"

	"github.com/ki4jnq/forge/deploy/shippers"
	"github.com/ki4jnq/forge/deploy/shippers/k8"
)

var ErrNotAShipper = errors.New("Undefined shipper requested in Forgefile")

type Ctx context.Context

type Config map[string]shipperBlock

type shipperBlock struct {
	ShipperName string `yaml:"shipper"`
	Opts        map[string]interface{}
}

func (sb *shipperBlock) toShipper() Shipper {
	switch sb.ShipperName {
	case "null-shipper":
		return shippers.NullShipper{}
	case "gulp-s3":
		return &shippers.GulpS3Shipper{}
	case "s3-copy":
		return &shippers.S3Copy{}
	case "k8":
		return k8.NewK8Shipper(sb.Opts)
	case "shell":
		return &shippers.ShellShipper{Opts: sb.Opts}
	case "app-engine":
		// args holds extra commandline arguments.
		return shippers.NewAppEngineShipper(sb.Opts, args)
	default:
		panic(ErrNotAShipper)
	}
}
