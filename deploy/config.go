package deploy

import (
	"errors"

	"github.com/ki4jnq/forge/deploy/engine"
	"github.com/ki4jnq/forge/deploy/shippers"
	"github.com/ki4jnq/forge/deploy/shippers/k8"
)

var ErrNotAShipper = errors.New("Undefined shipper requested in Forgefile")

type Config map[string]shipperBlock

// shipperBlock represents a configuration block from the Forgefile for an
// individual shipper object.
type shipperBlock struct {
	ShipperName string `yaml:"shipper"`
	Opts        map[string]interface{}
}

// toShipper builds a Shipper object from the configuration.
func (sb *shipperBlock) toShipper() engine.Shipper {
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
		return shippers.NewAppEngineShipper(sb.Opts)
	default:
		panic(ErrNotAShipper)
	}
}
