package cmd

import (
	"fmt"
)

// VersionCmd shows version information
type VersionCmd struct{}

func (v *VersionCmd) Run(name, version string) error {
	fmt.Printf("%s %s\n", APP_NAME, version)
	return nil
}
