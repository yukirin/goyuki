package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func TestRunCommand_implement(t *testing.T) {
	var _ cli.Command = &RunCommand{}
}

func TestRunCommandFlag(t *testing.T) {
	testCases := []struct {
		args   []string
		code   int
		result string
	}{
		{args: []string{"-foo", "lang", "-hoge"}, code: ExitCodeFailed, result: "Invalid option:"},
		{args: []string{}, code: ExitCodeFailed, result: "Invalid arguments"},
		{args: []string{"-l", "lang", "testdata/337", "foo.go"}, code: ExitCodeFailed, result: "Invalid language"},
		{args: []string{"-V", "hoge", "testdata/337", "foo.go"}, code: ExitCodeFailed, result: "Invalid Validater"},
	}

	for _, testCase := range testCases {
		ui := new(cli.MockUi)
		c := &RunCommand{
			Meta: Meta{
				UI: ui,
			},
		}

		code := c.Run(testCase.args)
		errs := ui.ErrorWriter.String()

		if code != testCase.code || !strings.Contains(errs, testCase.result) {
			t.Errorf("bad status code = %v; want %v\nError message = %s; want %s", code, testCase.code, errs, testCase.result)
		}
	}
}
