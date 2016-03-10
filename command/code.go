package command

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"
)

// Code is to compile and run the code
type Code struct {
	*LangCmd
	*Info
	Lang []string
	Dir  string
}

// Compile to compile the code
func (c *Code) Compile(r io.Reader, w, e io.Writer) error {
	cmd, err := c.buildCmd(0)
	if err != nil {
		return err
	}

	cmd.Stdin, cmd.Stdout, cmd.Stderr = r, w, e
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", CE, c.File)
	}
	return nil
}

func (c *Code) buildCmd(i int, args ...string) (*exec.Cmd, error) {
	var b bytes.Buffer
	t := template.Must(template.New("cmd").Parse(c.Lang[i]))
	if err := t.Execute(&b, c); err != nil {
		msg := fmt.Sprintf("executing template: %v", err)
		return nil, fmt.Errorf(msg)
	}

	com := strings.Split(b.String(), " ")
	com = append(com, args...)
	cmd := exec.Command(com[0], com[1:]...)
	cmd.Dir = c.Dir
	return cmd, nil
}

// Run to run the normal format of Judge
func (c *Code) Run(v Validater, expected []byte, r io.Reader, w, e io.Writer) (string, error) {
	cmd, err := c.buildCmd(1)
	if err != nil {
		return "", err
	}

	cmd.Stdin, cmd.Stdout, cmd.Stderr = r, w, e
	return c.judge(cmd, v, expected), nil
}

func (c *Code) judge(cmd *exec.Cmd, v Validater, expected []byte) string {
	ch := make(chan error)
	sTime := time.Now()
	go func() {
		ch <- cmd.Run()
	}()

	select {
	case err := <-ch:
		t := time.Now().Sub(sTime).Nanoseconds() / 1000000
		if err != nil {
			return RE
		}
		if !v.Validate(cmd.Stdout.(*bytes.Buffer).Bytes(), expected) {
			return fmt.Sprintf("%s: %d ms", WA, t)
		}
		return fmt.Sprintf("%s: %d ms", AC, t)
	case <-time.After(time.Duration(c.Info.Time) * time.Second):
		return TLE
	}
}

// Reactive to run the reactive format of Judge
func (c *Code) Reactive(code *Code, inFile, outFile string, r io.Reader, w, e io.Writer) (string, error) {
	cur, err := os.Getwd()
	if err != nil {
		return "", err
	}

	rCmd, err := c.buildCmd(1, cur+"/"+inFile, cur+"/"+outFile, code.Dir+"/"+code.LangCmd.File)
	if err != nil {
		return "", err
	}

	cmd, err := code.buildCmd(1)
	if err != nil {
		return "", err
	}

	rCmd.Stdin, err = cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if c.Info.JudgeType == Reactive {
		cmd.Stdin, err = rCmd.StdoutPipe()
		if err != nil {
			return "", err
		}
	} else {
		cmd.Stdin, rCmd.Stdout = r, w
	}
	cmd.Stderr, rCmd.Stderr = e, e
	return c.reactiveJudge(cmd, rCmd), nil
}

func (c *Code) reactiveJudge(cmd, rCmd *exec.Cmd) string {
	ch1, ch2 := make(chan error), make(chan []error)
	sTime := time.Now()
	go func() {
		ch1 <- cmd.Run()
	}()
	go func() {
		err := rCmd.Run()
		ch2 <- []error{<-ch1, err}
	}()

	select {
	case errs := <-ch2:
		t := time.Now().Sub(sTime).Nanoseconds() / 1000000
		if errs[0] != nil {
			return RE
		}
		if errs[1] != nil {
			return fmt.Sprintf("%s: %d ms", WA, t)
		}
		return fmt.Sprintf("%s: %d ms", AC, t)
	case <-time.After(time.Duration(c.Info.Time) * time.Second):
		return TLE
	}
}

// GetCode to get the compiled code
func GetCode(f string, lang []string, i *Info, w, e io.Writer) (*Code, *Result, func(), error) {
	dir, err := ioutil.TempDir("", "goyuki")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't create directory: %v", err)
	}
	clearFunc := func() {
		os.RemoveAll(dir)
	}

	ext := Ext(lang[2])
	_, source := path.Split(f)
	lCmd := LangCmd{
		File: source,
		Exec: strings.Split(source, ".")[0],
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read source file: %v", err)
	}

	if err := ioutil.WriteFile(dir+"/"+source, b, FPerm); err != nil {
		return nil, nil, nil, err
	}

	ret := &Result{
		info:       i,
		date:       time.Now(),
		lang:       lang[2],
		codeLength: len(b),
	}

	code := &Code{
		LangCmd: &lCmd,
		Lang:    lang,
		Info:    i,
		Dir:     dir,
	}

	sTime := time.Now()
	err = code.Compile(os.Stdin, w, e)
	ret.compileTime = time.Now().Sub(sTime)
	if err != nil {
		return nil, nil, nil, err
	}

	if ext[1:] == "java" || ext[1:] == "scala" {
		lCmd.Class, err = classFile(dir, source, ext[1:])
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return code, ret, clearFunc, nil
}

// GetReactiveCode to get the compiled reactive code
func GetReactiveCode(info *Info, no string, w, e io.Writer) (*Code, func(), error) {
	dir, err := ioutil.TempDir("", "goyuki")
	if err != nil {
		return nil, nil, fmt.Errorf("can't create directory: %v", err)
	}
	clearFunc := func() {
		os.RemoveAll(dir)
	}

	ext := Ext(info.RLang)
	lang, source := Lang[ext[1:]], ReactiveCode+ext
	lCmd := LangCmd{
		File: source,
		Exec: strings.Split(source, ".")[0],
	}

	b, err := ioutil.ReadFile(no + "/" + source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read reactive file: %v", err)
	}

	if ext[1:] == "java" {
		source = className(b) + ext
		lCmd.File = source
	}

	if err := ioutil.WriteFile(dir+"/"+source, b, FPerm); err != nil {
		return nil, nil, err
	}

	code := &Code{
		LangCmd: &lCmd,
		Lang:    lang,
		Info:    info,
		Dir:     dir,
	}

	if err := code.Compile(os.Stdin, w, e); err != nil {
		return nil, nil, err
	}
	if ext[1:] == "java" || ext[1:] == "scala" {
		lCmd.Class, err = classFile(dir, source, ext[1:])
		if err != nil {
			return nil, nil, err
		}
	}
	return code, clearFunc, nil
}
