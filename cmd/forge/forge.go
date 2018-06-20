package main

import (
	"fmt"
	"os"

	"github.com/ki4jnq/forge"
	_ "github.com/ki4jnq/forge/db"
	_ "github.com/ki4jnq/forge/deploy"
	_ "github.com/ki4jnq/forge/run"
	_ "github.com/ki4jnq/forge/version"
)

const helpMe = `Usage of forge:

  forge cmd [options]

Where "cmd" is one of:

  deploy
  run
  version
  db

To see the available options for some subcommand "cmd", run one of the following:

  forge cmd -h
  forge cmd --help
`

func main() {
	defer func() {
		if err := recover(); err != nil {
			prettyExit(err)
		}
	}()

	if len(os.Args) < 2 {
		fmt.Println(helpMe)
		prettyExit("ERROR: Not enough arguments")
	}
	cmdName := os.Args[1]

	cmd, err := forge.CmdForName(cmdName)
	if err != nil {
		panic(err)
	}

	if err = cmd.Run(); err != nil {
		prettyExit(err)
	}
}

func prettyExit(message interface{}) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
