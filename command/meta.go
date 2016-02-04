package command

import (
	"bufio"
	"crypto/rand"
	"encoding/base32"
	"flag"
	"io"
	"os"
	"strings"
)
import (
	"github.com/mgutz/ansi"
	"github.com/mitchellh/cli"
)

// Meta contain the meta-option that nearly all subcommand inherited.
type Meta struct {
	UI cli.Ui
}

// LangCmd is struct to fill Lang template
type LangCmd struct {
	File  string
	Exec  string
	Class string
}

// InfoFile is yukicoder problem infomation file
const InfoFile = "info.json"

// Dir of File permittion
const (
	DPerm = 0755
	FPerm = 0644
)

// Lang is compile and exec command template
var Lang = map[string][]string{
	"cpp":   {"g++ -O2 -lm -std=gnu++11 -o a.out {{.File}}", "./a.out"},
	"go":    {"go build {{.File}}", "./{{.Exec}}"},
	"c":     {"gcc -O2 -lm -o a.out {{.File}}", "./a.out"},
	"rb":    {"ruby --dissable-gems -w -c {{.File}}", "ruby --disable-gems {{.File}}"},
	"py2":   {"python2 -m py_compile {{.File}}", "python2 {{.Exec}}.pyc"},
	"py":    {"python3 -mpy_compile {{.File}}", "python3 {{.File}}"},
	"pypy2": {"pypy2 -m py_compile {{.File}}", "pypy2 {{.File}}"},
	"pypy3": {"pypy3 -mpy_compile {{.File}}", "pypy3 {{.File}}"},
	"js":    {"echo", "node {{.File}}"},
	"java":  {"javac -encoding UTF8 {{.File}}", "java -ea -Xmx700m -Xverify:none -XX:+TieredCompilation -XX:TieredStopAtLevel=1 {{.Class}}"},
	"pl":    {"perl -cw {{.File}}", "perl -X {{.File}}"},
	"pl6":   {"perl6 -cw {{.File}}", "perl6 {{.File}}"},
	"php":   {"php -l {{.File}}", "php {{.File}}"},
	"rs":    {"rustc {{.File}} -o Main", "./Main"},
	"scala": {"scalac {{.File}}", "scala {{.Class}}"},
	"hs":    {"ghc -o a.out -O {{.File}}", "./a.out"},
	"scm":   {"echo", "gosh {{.File}}"},
	"sh":    {"echo", "sh {{.File}}"},
	"txt":   {"echo", "cat {{.File}}"},
	"ml":    {"ocamlc str.cma {{.File}} -o a.out", "./a.out"},
}

// yukicoder Judge Code
var (
	AC  = ansi.Color("AC", "green+bh")
	WA  = ansi.Color("WA", "yellow+bh")
	TLE = ansi.Color("TLE", "yellow+bh")
	MLE = ansi.Color("MLE", "yellow+bh")
	RE  = ansi.Color("RE", "yellow+bh")
	CE  = ansi.Color("CE", "yellow+bh")
)

// NewFlagSet generates common flag.FlagSet
// https://github.com/tcnksm/gcli/blob/master/command/meta.go
func (m *Meta) NewFlagSet(name string, helpText string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)

	flags.Usage = func() { m.UI.Output(helpText) }

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

// TmpFile create random file name
func TmpFile() (string, error) {
	b := make([]byte, 25)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "="), nil
}

// MkTmpDir make tmp directory
func MkTmpDir() (string, error) {
	tmpDir := os.TempDir()
	dir, err := TmpFile()
	if err != nil {
		return "", err
	}
	tmpDir += "/" + dir

	err = os.Mkdir(tmpDir, DPerm)
	if err != nil {
		return "", err
	}

	return tmpDir, nil
}
