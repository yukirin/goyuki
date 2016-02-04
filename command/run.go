package command

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mgutz/ansi"
)

// RunCommand is a Command that run the test
type RunCommand struct {
	Meta
}

// Validater is the interface that wraps the Validate method
type Validater interface {
	Validate(actual, expected []byte) bool
}

// DiffValidater is verifies the exact match
type DiffValidater struct {
}

// Validate is verifies the exact match
func (d *DiffValidater) Validate(actual, expected []byte) bool {
	return bytes.Equal(actual, expected)
}

var validaters = map[string]Validater{
	"diff": &DiffValidater{},
}

var lang = map[string][][]string{
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
	"perl6": {{"perl6", "-cw", "__filename__"}, {"perl6", "__filename__"}},
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

// Run run the test
func (c *RunCommand) Run(args []string) int {
	var (
		langFlag     string
		validateFlag string
	)

	args, err := parseArgs([]*string{&langFlag, &validateFlag}, args)
	if err != nil {
		c.Ui.Error(fmt.Sprint(err))
		return 1
	}

	if len(args) < 2 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.Ui.Error(msg)
		return 1
	}

	if _, err := os.Stat(args[0]); err != nil {
		c.Ui.Error("does not exist (No such directory)")
		return 1
	}

	tmpDirName, err := mkTmpDir()
	if err != nil {
		c.Ui.Error(fmt.Sprint(err))
		return 1
	}
	defer os.RemoveAll(tmpDirName)

	_, source := path.Split(args[1])
	ext := path.Ext(args[1])[1:]
	if langFlag != "" {
		ext = langFlag
	}
	v := validaters["diff"]
	if validateFlag != "" {
		v = validaters[validateFlag]
	}

	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		msg := fmt.Sprintf("failed to read source file : %v", err)
		c.Ui.Error(msg)
		return 1
	}

	err = ioutil.WriteFile(tmpDirName+"/"+source, b, FPerm)
	if err != nil {
		c.Ui.Error(fmt.Sprint(err))
		return 1
	}

	if err := compile(lang[ext][0], source, tmpDirName); err != nil {
		c.Ui.Output(fmt.Sprint(err))
		return 1
	}

	class := ""
	if ext == "java" || ext == "scala" {
		class, err = classFile(tmpDirName)
		if err != nil {
			c.Ui.Error(fmt.Sprint(err))
			return 1
		}
	}

	infoBuf, err := ioutil.ReadFile(args[0] + "/" + "info.json")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to read info file : %v", err))
		return 1
	}

	info := Info{}
	if err := json.Unmarshal(infoBuf, &info); err != nil {
		c.Ui.Error(fmt.Sprint(err))
	}

	inFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_in", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("input testcase error : %v", err)
		c.Ui.Error(msg)
		return 1
	}

	outFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_out", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("output testcase error : %v", err)
		c.Ui.Error(msg)
		return 1
	}

	for i := 0; i < len(inFiles); i++ {
		err := func() error {
			var execCom []string
			for _, s := range lang[ext][1] {
				s = strings.Replace(s, "__filename__", source, 1)
				s = strings.Replace(s, "__class__", class, 1)
				s = strings.Replace(s, "__exec__", strings.Replace(source, path.Ext(source), "", 1), 1)
				execCom = append(execCom, s)
			}

			cmd := exec.Command(execCom[0], execCom[1:]...)
			cmd.Dir = tmpDirName

			input, err := os.Open(inFiles[i])
			if err != nil {
				msg := fmt.Sprintf("input test file error : %v", err)
				c.Ui.Error(msg)
				return err
			}
			defer input.Close()

			output, err := ioutil.ReadFile(outFiles[i])
			if err != nil {
				msg := fmt.Sprintf("output test file error : %v", err)
				c.Ui.Error(msg)
				return err
			}

			var buf bytes.Buffer
			cmd.Stdin = input
			cmd.Stdout = &buf
			cmd.Stderr = os.Stderr

			result := judge(cmd, output, v, &info)
			c.Ui.Output(result)
			return nil
		}()
		if err != nil {
			return 1
		}
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

Usage:
	goyuki run problem_no exec_file

`
	return strings.TrimSpace(helpText)
}

func parseArgs(sp []*string, args []string) ([]string, error) {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	flags.Usage = func() {}
	flags.StringVar(sp[0], "l", "", "Specify Language")
	flags.StringVar(sp[1], "validate", "", "Specify Validater")

	if err := flags.Parse(args); err != nil {
		return nil, fmt.Errorf("Invalid option: %s", strings.Join(args, " "))
	}
	return flags.Args(), nil
}

func compile(cmds []string, file, tmpDir string) error {
	var compileCom []string
	for _, s := range cmds {
		s = strings.Replace(s, "__filename__", file, 1)
		compileCom = append(compileCom, s)
	}

	cmd := exec.Command(compileCom[0], compileCom[1:]...)
	cmd.Dir = tmpDir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s (Compile Error)", CE)
	}
	return nil
}

func judge(cmd *exec.Cmd, expected []byte, v Validater, i *Info) string {
	ch := make(chan error)
	sTime := time.Now()

	go func() {
		ch <- cmd.Run()
	}()

	select {
	case err := <-ch:
		t := time.Now().Sub(sTime).Nanoseconds() / 1000000

		if err != nil {
			return fmt.Sprintf("%s (Runtime Error)", RE)
		}

		if !v.Validate(cmd.Stdout.(*bytes.Buffer).Bytes(), expected) {
			return fmt.Sprintf("%s (Wrong Answer) : %d ms", WA, t)
		}

		return fmt.Sprintf("%s (Accepted) : %d ms", AC, t)
	case <-time.After(time.Duration(i.Time) * time.Second):
		return fmt.Sprintf("%s (Time Limit Exceeded)", TLE)
	}
}

func tmpFile() (string, error) {
	b := make([]byte, 25)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "="), nil
}

func mkTmpDir() (string, error) {
	tmpDir := os.TempDir()
	dir, err := tmpFile()
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

func classFile(dir string) (string, error) {
	files, err := filepath.Glob(dir + "/*")
	if err != nil {
		return "", err
	}
	for _, s := range files {
		if strings.HasSuffix(s, "$.class") {
			continue
		}
		if strings.HasSuffix(s, ".class") {
			_, f := path.Split(s)
			return strings.TrimSuffix(f, ".class"), nil
		}
	}
	return "", errors.New("missing .class file")
}
