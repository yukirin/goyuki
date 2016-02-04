package command

import (
	"bufio"
	"flag"
	"io"
)
import "github.com/mitchellh/cli"

// Meta contain the meta-option that nearly all subcommand inherited.
type Meta struct {
	UI cli.Ui
}

// NewFlagSet generates common flag.FlagSet
// https://github.com/tcnksm/gcli/blob/master/command/meta.go
func (m *Meta) NewFlagSet(name string, helpText string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	flags.Usage = func() { m.UI.Error(helpText) }

	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	flags.SetOutput(errW)

	go func() {
		for errScanner.Scan() {
			m.UI.Error(errScanner.Text())
		}
	}()

	return flags
}
