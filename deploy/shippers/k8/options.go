package k8

import (
	"fmt"
)

type Options map[string]interface{}

func (o Options) mustLookup(key string) string {
	name, ok := o[key].(string)
	if !ok {
		panic(ConfigErr{key})
	}
	return name
}

// readConfigInto reads options from the Forge config into variables passed
// in the `targets` varargs. Care should be taken to ensure that len(optNames)
// is <= len(targets).
func (o Options) readConfigInto(optNames []string, targets ...interface{}) error {
	for idx, name := range optNames {
		if _, ok := o[name]; !ok {
			return ErrMissingConfig
		}

		value, ok := o[name].(string)
		if !ok {
			return ErrConfigInvalid
		}

		switch target := targets[idx].(type) {
		case *string:
			*target = value
			continue
		case *[]byte:
			*target = []byte(value)
			continue
		default:
			fmt.Println("failed to match %v\n", targets[idx])
		}
		return ErrConfigInvalid
	}
	return nil
}
