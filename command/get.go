package command

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"strings"
)

// GetCommand is a Command that get test case
type GetCommand struct {
	Meta
}

var config = "~/.goyuki"

// Run get test case
func (c *GetCommand) Run(args []string) int {
	if len(args) < 1 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.Ui.Error(msg)
		return 1
	}

	cookie, err := readCookie(config)
	if err != nil {
		c.Ui.Error(fmt.Sprint(err))
		return 1
	}
	fmt.Println(cookie)

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

func readCookie(config string) (string, error) {
	cookie := ""
	usr, err := user.Current()
	if err != nil {
		return cookie, err
	}

	b, err := ioutil.ReadFile(strings.Replace(config, "~", usr.HomeDir, 1))
	if err != nil {
		return cookie, err
	}

	return strings.Trim(string(b), "\n"), nil
}
