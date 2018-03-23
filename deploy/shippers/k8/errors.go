package k8

import (
	"fmt"
)

type ConfigErr struct {
	opt string
}

func (mce ConfigErr) Error() string {
	return fmt.Sprintf("Error while looking up option \"%v\"\n", mce.opt)
}
