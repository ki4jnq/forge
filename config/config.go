package config

import (
	"flag"
	"github.com/ki4jnq/rconf"
	"os"
)

type Config struct {
	Env string
}

type Cmd struct {
	Name    string
	SubConf interface{}

	Flags    *flag.FlagSet
	RubyConf func(rconf.Binder)

	SubRunner func() error
}

func (cmd *Cmd) Run() error {
	cmd.Flags.Parse(os.Args[2:])
	//ParseConfig(cmd.Conf.Env)
	ParseRbConfig(&Conf)

	return cmd.SubRunner()
}
