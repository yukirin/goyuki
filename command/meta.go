package command

import (
	"bufio"
	"flag"
	"io"

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

// BaseURL is yukicoder problem url
const BaseURL = "http://yukicoder.me/problems"

// InfoFile is yukicoder problem infomation file
const InfoFile = "info.json"

// ReactiveCode is reactive judge code file
const ReactiveCode = "reactive"

// ExitCodes
const (
	ExitCodeOK = iota
	ExitCodeFailed
)

// Judge Type
const (
	Normal = iota
	Special
	Reactive
)

// Dir and File permittion
const (
	DPerm = 0755
	FPerm = 0644
)

// Lang is compile and exec command template
var Lang = map[string][]string{
	"cpp":   {"g++ -O2 -lm -std=gnu++11 -o a.out {{.File}}", "./a.out", "C++11"},
	"go":    {"go build {{.File}}", "./{{.Exec}}", "Go"},
	"c":     {"gcc -O2 -lm -o a.out {{.File}}", "./a.out", "C"},
	"rb":    {"ruby --disable-gems -w -c {{.File}}", "ruby --disable-gems {{.File}}", "Ruby"},
	"py2":   {"python2 -m py_compile {{.File}}", "python2 {{.Exec}}.pyc", "Python2"},
	"py":    {"python3 -mpy_compile {{.File}}", "python3 {{.File}}", "Python3"},
	"pypy2": {"pypy -m py_compile {{.File}}", "pypy {{.File}}", "PyPy2"},
	"pypy3": {"pypy3 -mpy_compile {{.File}}", "pypy3 {{.File}}", "PyPy3"},
	"js":    {"echo", "node {{.File}}", "JavaScript"},
	"java":  {"javac -d ./ -encoding UTF8 {{.File}}", "java -ea -Xmx700m -Xverify:none -XX:+TieredCompilation -XX:TieredStopAtLevel=1 {{.Class}}", "Java8"},
	"pl":    {"perl -cw {{.File}}", "perl -X {{.File}}", "Perl"},
	"pl6":   {"perl6 -cw {{.File}}", "perl6 {{.File}}", "Perl6"},
	"php":   {"php -l {{.File}}", "php {{.File}}", "PHP"},
	"rs":    {"rustc {{.File}} -o Main", "./Main", "Rust"},
	"scala": {"scalac {{.File}}", "scala {{.Class}}", "Scala"},
	"hs":    {"ghc -o a.out -O {{.File}}", "./a.out", "Haskell"},
	"scm":   {"echo", "gosh {{.File}}", "Scheme"},
	"sh":    {"echo", "sh {{.File}}", "Bash"},
	"txt":   {"echo", "cat {{.File}}", "Text"},
	"ml":    {"ocamlc str.cma {{.File}} -o a.out", "./a.out", "OCaml"},
	"cs":    {"dmcs -warn:0 /r:System.Numerics.dll /codepage:utf8 {{.File}} -out:a.exe", "mono a.exe", "C#"},
	"d":     {"dmd -m64 -w -wi -O -release -inline -I/usr/include/dmd/druntime/import/ -I/usr/include/dmd/phobos -ofa.out {{.File}}", "./a.out", "D"},
	"nim":   {"nim --hints:off -o:a.out -d:release c {{.File}}", "./a.out", "Nim"},
	"kt":    {"kotlinc {{.File}} -include-runtime -d main.jar", "java -jar main.jar", "Kotlin"},
	"cr":    {"crystal build -o a.out --release {{.File}}", "./a.out", "Crystal"},
	"fs":    {"fsharpc {{.File}} -o ./a.exe", "./a.exe", "F#"},
	"f90":   {"gfortran {{.File}} -o ./a.out", "./a.out", "Fortran"},
}

// yukicoder Judge Code
var (
	AC  = ansi.Color("[AC]", "green+bh")
	WA  = ansi.Color("[WA]", "yellow+bh")
	TLE = ansi.Color("[TLE]", "yellow+bh")
	MLE = ansi.Color("[MLE]", "yellow+bh")
	RE  = ansi.Color("[RE]", "yellow+bh")
	CE  = ansi.Color("[CE]", "yellow+bh")
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

// Ext to get an extension from the language
func Ext(lang string) string {
	for k, v := range Lang {
		if lang != v[2] {
			continue
		}
		return "." + k
	}
	return ""
}
