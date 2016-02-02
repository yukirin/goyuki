package command

import (
	"strings"
)

// RunCommand is a Command that run the test
type RunCommand struct {
	Meta
}

// Run run the test
func (c *RunCommand) Run(args []string) int {
	// Write your code here

	return 0
}

// Synopsis is a one-line, short synopsis of the command.
func (c *RunCommand) Synopsis() string {
	return "Run the test"
}

// Help is a long-form help text
func (c *RunCommand) Help() string {
	helpText := `
Usage:
	goyuki run problem_no exec_file

`
	return strings.TrimSpace(helpText)
}
