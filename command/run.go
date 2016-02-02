package command

import (
	"strings"
)

type RunCommand struct {
	Meta
}

func (c *RunCommand) Run(args []string) int {
	// Write your code here

	return 0
}

func (c *RunCommand) Synopsis() string {
	return ""
}

func (c *RunCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
