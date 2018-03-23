package forge

import (
	"bytes"
	"errors"
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

var (
	Registry = make(map[string]*Cmd)

	ErrNoCmd = errors.New("Sub Command does not exist")

	forgefile = "Forgefile"
)

// Register registers subcommands and their configurations in a central
// location.
func Register(cmd *Cmd) {
	cmd.Conf = &Config{}
	cmd.Flags.StringVar(&cmd.Conf.Env, "env", "development", "Set the environment for forge to run in.")

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

func ParseConfig(env string) {
	buffer := &bytes.Buffer{}
	tmpl := template.Must(
		template.New(
			"forgefile",
		).Funcs(template.FuncMap{
			"env": os.Getenv,
			"def": defaultValue,
		}).ParseFiles(forgefile),
	).Lookup("Forgefile")

	if err := tmpl.Execute(buffer, struct{}{}); err != nil {
		panic(err)
	}

	unformatter := NewParser(env)
	if err := yaml.Unmarshal(buffer.Bytes(), unformatter); err != nil {
		panic(err)
	}
}

func defaultValue(defaultVal, val string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
