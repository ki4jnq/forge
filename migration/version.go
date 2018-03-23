package migration

import (
	"fmt"
)

type Versioner interface {
	VersionParts() (int, int, int)
	String() string
}

// This object could probably be abstracted into it's own package, allowing
// for reuse in the `version` sub-command.
type Version struct {
	Major int
	Minor int
	Patch int
}

func VersionFromString(version string) (Version, error) {
	v := Version{}
	if err := v.Scan(version); err != nil {
		return Version{}, err
	}
	return v, nil
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *Version) Scan(input string) error {
	return v.Scanf(input, "%d.%d.%d")
}

func (v *Version) Scanf(input, format string) error {
	_, err := fmt.Sscanf(input, format, &v.Major, &v.Minor, &v.Patch)
	return err
}

func (v *Version) VersionParts() (int, int, int) {
	return v.Major, v.Minor, v.Patch
}

// Compare compares this Version to "other" to determine which comes first. It
// returns 1 if other is less than, 0 if it is equal, -1 if it is greater than.
func (v *Version) Compare(other Versioner) int {
	oMaj, oMin, oPat := other.VersionParts()
	if v.Major < oMaj {
		return -1
	} else if v.Major > oMaj {
		return 1
	} else if v.Minor < oMin {
		return -1
	} else if v.Minor > oMin {
		return 1
	} else if v.Patch < oPat {
		return -1
	} else if v.Patch > oPat {
		return 1
	}
	return 0
}

func (v *Version) IsZero() bool {
	if v.Major+v.Minor+v.Patch == 0 {
		return true
	}
	return false
}
