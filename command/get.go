package command

import (
	"strings"
)

type GetCommand struct {
	Meta
}

func (c *GetCommand) Run(args []string) int {
	// Write your code here

	return 0
}

func (c *GetCommand) Synopsis() string {
	return ""
}

func (c *GetCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
