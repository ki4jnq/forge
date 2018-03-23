package migration

import (
	"errors"
	"fmt"
)

var (
	FinalMigration = errors.New("There are no more migrations.")
)

type VersionNotFound struct {
	version Version
}

func (e VersionNotFound) Error() string {
	return fmt.Sprintf("No SQL file found for %v\n", e.version.String())
}
