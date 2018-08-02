package config

import (
	"errors"
)

var (
	Conf     = Config{}
	Registry = make(map[string]*Cmd)
	ErrNoCmd = errors.New("Sub Command does not exist")
)

// Register registers subcommands and their configurations in a central
// location.
func Register(cmd *Cmd) {
	cmd.Flags.StringVar(
		&Conf.Env,
		"env",
		"development",
		"Set the environment for forge to run in.",
	)

	Registry[cmd.Name] = cmd
}

func IsRegisteredCmd(cmdName string) bool {
	_, ok := Registry[cmdName]
	return ok
}

func CmdForName(name string) (*Cmd, error) {
	cmd, ok := Registry[name]
	if ok {
		return cmd, nil
	}
	return nil, ErrNoCmd
}
