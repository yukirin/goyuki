package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
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
