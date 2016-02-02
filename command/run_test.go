package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestRunCommand_implement(t *testing.T) {
	var _ cli.Command = &RunCommand{}
}
