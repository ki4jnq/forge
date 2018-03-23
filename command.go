package forge

import (
	"flag"
	"os"
)

type Cmd struct {
	Name    string
	Conf    *Config
	SubConf interface{}
	Flags   *flag.FlagSet

	SubRunner func() error
}

func (cmd *Cmd) Run() error {
	cmd.Flags.Parse(os.Args[2:])
	ParseConfig(cmd.Conf.Env)

	return cmd.SubRunner()
}
