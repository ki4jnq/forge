package shippers

import (
	"errors"
	"fmt"
)

// failSafe should be used in conjunction with defer to ensure that any errors
// encountered during shipping a service are properly handled and logged.
func failSafe(preamble string, ch chan error) {
	if err := recover(); err != nil {
		if e, ok := err.(error); ok {
			ch <- e
		} else {
			ch <- errors.New(fmt.Sprintf("%s: %v\n", preamble, err))
		}
	}
}
