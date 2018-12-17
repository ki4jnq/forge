package deploy

import (
	"flag"

	"github.com/ki4jnq/forge"
	"github.com/ki4jnq/forge/deploy/engine"
)

var (
	conf  = Config{}
	flags = flag.NewFlagSet("deploy", flag.ExitOnError)
	opts  = engine.Options{}
)

func init() {
	flags.StringVar(
		&opts.AppEngine.ImageTag,
		"ae-image-tag",
		"",
		"The AppEngine Docker image tag to deploy",
	)
	flags.StringVar(
		&opts.Version,
		"version",
		"",
		"The version number to deploy.",
	)

	forge.Register(&forge.Cmd{
		Name:      "deploy",
		Flags:     flags,
		SubConf:   conf,
		SubRunner: run,
	})
}

func run() error {
	shippers := make(map[string]engine.Shipper, len(conf))
	for target, block := range conf {
		shippers[target] = block.toShipper()
	}

	eng := engine.NewEngine(shippers)
	return eng.Run(opts)
}
