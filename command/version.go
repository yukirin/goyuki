package command

import (
	"bytes"
	"fmt"
)

// VersionCommand is a Command that shows version
type VersionCommand struct {
	Meta

	Name     string
	Version  string
	Revision string
}

// Run shows version string and commit hash if it exists.
func (c *VersionCommand) Run(args []string) int {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "%s version %s", c.Name, c.Version)
	if c.Revision != "" {
		fmt.Fprintf(&versionString, " (%s)", c.Revision)
	}

	c.UI.Output(versionString.String())
	return 0
}

// Synopsis is a one-line, short synopsis of the command.
func (c *VersionCommand) Synopsis() string {
	return fmt.Sprintf("Print %s version and quit", c.Name)
}

// Help is a long-form help text.
func (c *VersionCommand) Help() string {
	return ""
}
