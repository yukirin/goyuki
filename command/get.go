package command

import (
	"fmt"
	"strings"
)

// GetCommand is a Command that get test case
type GetCommand struct {
	Meta
}

// Run get test case
func (c *GetCommand) Run(args []string) int {
	if len(args) < 1 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.Ui.Error(msg)
		return 1
	}

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
