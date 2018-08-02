package config

import (
	"github.com/ki4jnq/rconf"
)

func ParseRbConfig(appConf *Config) {
	binder, exec, err := rconf.New()
	if err != nil {
		// TODO: Remove Panic
		panic(err)
	}

	binder.BlockWithArg("env", func(b rconf.Binder) {
		env := b.StringArgAt(0)
		if env != appConf.Env {
			// Don't try to configure anything for any environment other than
			// the one we are running in.
			return
		}

		for name, cmd := range Registry {
			if cmd.RubyConf != nil {
				b.Block(name, cmd.RubyConf)
			}
		}
	})

	exec("forge.rb")
}
