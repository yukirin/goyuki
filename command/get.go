package command

import (
	"strings"
)

// GetCommand is a Command that get test case
type GetCommand struct {
	Meta
}

// Run get test case
func (c *GetCommand) Run(args []string) int {
	// Write your code here

	return 0
}

// Synopsis is a one-line, short synopsis of the command.
func (c *GetCommand) Synopsis() string {
	return "Get test case"
}

// Help is a long-form help text
func (c *GetCommand) Help() string {
	helpText := `
Usage:
	goyuki get problem_no

`
	return strings.TrimSpace(helpText)
}
