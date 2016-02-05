package command

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/cli"
)

func TestGetCommand_implement(t *testing.T) {
	var _ cli.Command = &GetCommand{}
}

func TestGetCommand(t *testing.T) {
	ui := new(cli.MockUi)
	c := &GetCommand{
		Meta: Meta{
			UI: ui,
		},
	}

	tmpDir, err := ioutil.TempDir("", "test-get-command")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	backFunc, err := tmpChdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer backFunc()

	args := []string{"337"}

	if code := c.Run(args); code != ExitCodeOK {
		t.Fatalf("bad status code = %v; want %v\n\n%s", code, ExitCodeOK, ui.ErrorWriter.String())
	}

	if !equalFiles(prev+"/testdata/337", tmpDir+"/337") {
		t.Errorf("failed download testcase; problem no %s\n", args[0])
	}
}

func TestGetCommandEnvUnSet(t *testing.T) {
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
