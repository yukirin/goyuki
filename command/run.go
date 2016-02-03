package command

import (
	"fmt"
	"strings"
)

// RunCommand is a Command that run the test
type RunCommand struct {
	Meta
}

// Run run the test
func (c *RunCommand) Run(args []string) int {
	if len(args) < 2 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.Ui.Error(msg)
		return 1
	}

	return 0
}

// Synopsis is a one-line, short synopsis of the command.
func (c *RunCommand) Synopsis() string {
	return "テストを実行する"
}

// Help is a long-form help text
func (c *RunCommand) Help() string {
	helpText := `
problem_noで指定された番号の問題のテストを実行する
もしテストケースを取得していなければ、取得する

Usage:
	goyuki run problem_no exec_file

`
	return strings.TrimSpace(helpText)
}
