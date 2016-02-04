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

// Config is config file
const Config = "~/.goyuki"

// InfoFile is yukicoder problem infomation file
const InfoFile = "info.json"

// Dir of File permittion
const (
	DPerm = 0755
	FPerm = 0644
)

// Lang is compile and exec command
var Lang = map[string][][]string{
	"cpp":   {{"g++", "-O2", "-lm", "-std=gnu++11", "-o", "a.out", "__filename__"}, {"./a.out"}},
	"go":    {{"go", "build", "__filename__"}, {"./__exec__"}},
	"c":     {{"gcc", "-O2", "-lm", "-o", "a.out", "__filename__"}, {"./a.out"}},
	"rb":    {{"ruby", "--disable-gems", "-w", "-c", "__filename__"}, {"ruby", "--disable-gems", "__filename__"}},
	"py2":   {{"python2", "-m", "py_compile", "__filename__"}, {"python2", "__exec__.pyc"}},
	"py":    {{"python3", "-mpy_compile", "__filename__"}, {"python3", "__filename__"}},
	"pypy2": {{"pypy2", "-m", "py_compile", "__filename__"}, {"pypy2", "__filename__"}},
	"pypy3": {{"pypy3", "-mpy_compile", "__filename__"}, {"pypy3", "__filename__"}},
	"js":    {{"echo"}, {"node", "__filename__"}},
	"java":  {{"javac", "-encoding", "UTF8", "__filename__"}, {"java", "-ea", "-Xmx700m", "-Xverify:none", "-XX:+TieredCompilation", "-XX:TieredStopAtLevel=1", "__class__"}},
	"pl":    {{"perl", "-cw", "__filename__"}, {"perl", "-X", "__filename__"}},
	"pl6":   {{"perl6", "-cw", "__filename__"}, {"perl6", "__filename__"}},
	"php":   {{"php", "-l", "__filename__"}, {"php", "__filename__"}},
	"rs":    {{"rustc", "__filename__", "-o", "Main"}, {"./Main"}},
	"scala": {{"scalac", "__filename__"}, {"scala", "__class__"}},
	"hs":    {{"ghc", "-o", "a.out", "-O", "__filename__"}, {"./a.out"}},
	"scm":   {{"echo"}, {"gosh", "__filename__"}},
	"sh":    {{"echo"}, {"sh", "__filename__"}},
	"txt":   {{"echo"}, {"cat", "__filename__"}},
	"ml":    {{"ocamlc", "str.cma", "__filename__", "-o", "a.out"}, {"./a.out"}},
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
