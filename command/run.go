package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
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

// Run run the test
func (c *RunCommand) Run(args []string) int {
	var (
		langFlag      string
		validaterFlag string
	)

	flags := c.Meta.NewFlagSet("run", c.Help())
	flags.StringVar(&langFlag, "l", "", "Specify Language")
	flags.StringVar(&langFlag, "language", "", "Specify Language")
	flags.StringVar(&validaterFlag, "V", "", "Specify Validater")
	flags.StringVar(&validaterFlag, "validater", "", "Specify Validater")

	if err := flags.Parse(args); err != nil {
		msg := fmt.Sprintf("Invalid option: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return 1
	}
	args = flags.Args()

	if len(args) < 2 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return 1
	}

	if _, err := os.Stat(args[0]); err != nil {
		c.UI.Error("does not exist (No such directory)")
		return 1
	}

	tmpDir, err := MkTmpDir()
	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}
	defer os.RemoveAll(tmpDir)

	_, source := path.Split(args[1])
	ext := path.Ext(args[1])[1:]
	if langFlag != "" {
		ext = langFlag
	}
	v := validaters["diff"]
	if validaterFlag != "" {
		v = validaters[validaterFlag]
	}

	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		msg := fmt.Sprintf("failed to read source file : %v", err)
		c.UI.Error(msg)
		return 1
	}

	err = ioutil.WriteFile(tmpDir+"/"+source, b, FPerm)
	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	if err := compile(Lang[ext][0], source, tmpDir); err != nil {
		c.UI.Output(fmt.Sprint(err))
		return 1
	}

	class := ""
	if ext == "java" || ext == "scala" {
		class, err = classFile(tmpDir)
		if err != nil {
			c.UI.Error(fmt.Sprint(err))
			return 1
		}
	}

	infoBuf, err := ioutil.ReadFile(args[0] + "/" + "info.json")
	if err != nil {
		c.UI.Error(fmt.Sprintf("failed to read info file : %v", err))
		return 1
	}

	info := Info{}
	if err := json.Unmarshal(infoBuf, &info); err != nil {
		c.UI.Error(fmt.Sprint(err))
	}

	inputFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_in", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("input testcase error : %v", err)
		c.UI.Error(msg)
		return 1
	}

	outputFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_out", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("output testcase error : %v", err)
		c.UI.Error(msg)
		return 1
	}

	for i := 0; i < len(inputFiles); i++ {
		err := func() error {
			var execCom []string
			for _, s := range Lang[ext][1] {
				s = strings.Replace(s, "__filename__", source, 1)
				s = strings.Replace(s, "__class__", class, 1)
				s = strings.Replace(s, "__exec__", strings.Replace(source, path.Ext(source), "", 1), 1)
				execCom = append(execCom, s)
			}

			cmd := exec.Command(execCom[0], execCom[1:]...)
			cmd.Dir = tmpDir

			input, err := os.Open(inputFiles[i])
			if err != nil {
				msg := fmt.Sprintf("input test file error : %v", err)
				c.UI.Error(msg)
				return err
			}
			defer input.Close()

			output, err := ioutil.ReadFile(outputFiles[i])
			if err != nil {
				msg := fmt.Sprintf("output test file error : %v", err)
				c.UI.Error(msg)
				return err
			}

			var buf bytes.Buffer
			cmd.Stdin = input
			cmd.Stdout = &buf
			cmd.Stderr = os.Stderr

			result := judge(cmd, output, v, &info)
			c.UI.Output(result)
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
	return "コンパイル後、テストを実行する"
}

// Help is a long-form help text
func (c *RunCommand) Help() string {
	helpText := `
problem_noで指定された番号の問題のテストを実行する

Usage:
	goyuki run problem_no source_file

Options:
	-language, -l		実行する言語を指定します (デフォルト 拡張子から判別)
	-validater, -V       テストの一致方法を指定します (デフォルト diff validator)

`
	return strings.TrimSpace(helpText)
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
