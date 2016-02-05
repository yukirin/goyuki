package command

import (
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func TestGetCommand_implement(t *testing.T) {
	var _ cli.Command = &GetCommand{}
}

func TestGetCommandUnSetEnv(t *testing.T) {
	ui := new(cli.MockUi)
	c := &GetCommand{
		Meta: Meta{
			UI: ui,
		},
	}

	clearFunc := setEnv("GOYUKI", "")
	defer clearFunc()

	result := "$GOYUKI not set"
	args := []string{"337"}
	code := c.Run(args)
	errs := ui.ErrorWriter.String()

	if code != ExitCodeFailed || !strings.Contains(errs, result) {
		t.Errorf("bad status code = %v; want %v\nError message = %s; want %s", code, ExitCodeFailed, errs, result)
	}
}

func TestGetCommandWrongCookie(t *testing.T) {
	ui := new(cli.MockUi)
	c := &GetCommand{
		Meta: Meta{
			UI: ui,
		},
	}

	clearFunc := setEnv("GOYUKI", "test")
	defer clearFunc()

	result := "please log in to yukicoder"
	args := []string{"337"}
	code := c.Run(args)
	errs := ui.ErrorWriter.String()

	if code != ExitCodeFailed || !strings.Contains(errs, result) {
		t.Errorf("bad status code = %v; want %v\nError message = %s; want %s", code, ExitCodeFailed, errs, result)
	}
}

func TestGetCommandWrongProblem(t *testing.T) {
	ui := new(cli.MockUi)
	c := &GetCommand{
		Meta: Meta{
			UI: ui,
		},
	}

	clearFunc := setEnv("GOYUKI", "test")
	defer clearFunc()

	result := "the problem does not exist"
	args := []string{"99999"}
	code := c.Run(args)
	errs := ui.ErrorWriter.String()

	if code != ExitCodeFailed || !strings.Contains(errs, result) {
		t.Errorf("bad status code = %v; want %v\nError message = %s; want %s", code, ExitCodeFailed, errs, result)
	}
}

func TestGetCommandFlag(t *testing.T) {
	testCases := []struct {
		args   []string
		code   int
		result string
	}{
		{args: []string{"-l", "lang", "-hoge"}, code: ExitCodeFailed, result: "Invalid option"},
		{args: []string{}, code: ExitCodeFailed, result: "Invalid arguments"},
		{args: []string{"foobar"}, code: ExitCodeFailed, result: "invalid syntax"},
	}

	clearFunc := setEnv("GOYUKI", "test")
	defer clearFunc()

	for _, testCase := range testCases {
		ui := new(cli.MockUi)
		c := &GetCommand{
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
