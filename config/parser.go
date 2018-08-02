package config

import (
	"bytes"
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

var forgefile = "Forgefile"

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
