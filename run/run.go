package run

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ki4jnq/forge"
)

var (
	conf  = &Config{}
	flags = flag.NewFlagSet("run", flag.ExitOnError)
)

type Config struct {
	Env     map[string]string
	DefArgs map[string][]string `yaml:"args"`
}

func init() {
	forge.Register(&forge.Cmd{
		Name:      "run",
		Flags:     flags,
		SubConf:   conf,
		SubRunner: run,
	})
}

func run() error {
	// Get a list of all args that came after the first, non-matched cmdline
	// argument. These will get forwarded onto the sub process.
	var args []string
	var mainCmd string

	idx := conf.getPartitionIdx(flags.Args())
	if idx != -1 {
		args = flags.Args()[idx:]
		mainCmd = strings.Join(flags.Args()[0:idx], " ")
	} else {
		mainCmd = strings.Join(flags.Args()[0:], " ")
	}

	if defaults, ok := conf.DefArgs[mainCmd]; ok {
		args = append(defaults, args...)
	}

	subProc := exec.Command(mainCmd, args...)

	subProc.Stdin = os.Stdin
	subProc.Stdout = os.Stdout
	subProc.Stderr = os.Stderr

	conf.addEnvOpts(subProc)
	conf.addExternalEnv(subProc)

	err := subProc.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

func (c *Config) addEnvOpts(subProc *exec.Cmd) {
	subProc.Env = make([]string, len(c.Env))

	for name, value := range c.Env {
		pair := fmt.Sprintf("%v=%v", name, value)
		subProc.Env = append(subProc.Env, pair) // This is inefficient.
	}
}

func (c *Config) addExternalEnv(subProc *exec.Cmd) {
	subProc.Env = append(subProc.Env, os.Environ()...)
}

func (c *Config) getPartitionIdx(args []string) int {
	for idx, item := range args {
		if item == "--" {
			return idx
		}
	}
	return -1
}
