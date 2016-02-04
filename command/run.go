package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// RunCommand is a Command that run the test
type RunCommand struct {
	Meta
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

	if langFlag == "" {
		langFlag = strings.Replace(path.Ext(args[1]), ".", "", -1)
	}
	lang, ok := Lang[langFlag]
	if !ok {
		msg := fmt.Sprintf("Invalid language: %s", langFlag)
		c.UI.Error(msg)
		return 1
	}

	if validaterFlag == "" {
		validaterFlag = "diff"
	}
	v, ok := Validaters[validaterFlag]
	if !ok {
		msg := fmt.Sprintf("Invalid Validater: %s", validaterFlag)
		c.UI.Error(msg)
		return 1
	}

	lCmd := LangCmd{
		File: source,
		Exec: strings.Split(source, ".")[0],
	}

	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		msg := fmt.Sprintf("failed to read source file: %v", err)
		c.UI.Error(msg)
		return 1
	}

	err = ioutil.WriteFile(tmpDir+"/"+source, b, FPerm)
	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	infoBuf, err := ioutil.ReadFile(args[0] + "/" + "info.json")
	if err != nil {
		c.UI.Error(fmt.Sprintf("failed to read info file: %v", err))
		return 1
	}

	info := Info{}
	if err := json.Unmarshal(infoBuf, &info); err != nil {
		c.UI.Error(fmt.Sprint(err))
	}

	result := Result{
		info:       &info,
		date:       time.Now(),
		lang:       lang[2],
		codeLength: len(b),
	}

	sTime := time.Now()
	err = compile(lang[0], &lCmd, tmpDir)
	result.compileTime = time.Now().Sub(sTime)

	c.UI.Output(result.String())
	if err != nil {
		c.UI.Output(fmt.Sprint(err))
		return 1
	}

	if langFlag == "java" || langFlag == "scala" {
		class, err := classFile(tmpDir)
		lCmd.Class = class

		if err != nil {
			c.UI.Error(fmt.Sprint(err))
			return 1
		}
	}

	inputFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_in", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("input testcase error: %v", err)
		c.UI.Error(msg)
		return 1
	}

	outputFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_out", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("output testcase error: %v", err)
		c.UI.Error(msg)
		return 1
	}

	var execBuf bytes.Buffer

	t := template.Must(template.New("exec").Parse(lang[1]))
	if err := t.Execute(&execBuf, lCmd); err != nil {
		msg := fmt.Sprintf("executing template: %v", err)
		c.UI.Error(msg)
		return 1
	}

	execCom := strings.Split(execBuf.String(), " ")

	for i := 0; i < len(inputFiles); i++ {
		err := func() error {
			cmd := exec.Command(execCom[0], execCom[1:]...)
			cmd.Dir = tmpDir

			input, err := os.Open(inputFiles[i])
			if err != nil {
				msg := fmt.Sprintf("input test file error: %v", err)
				c.UI.Error(msg)
				return err
			}
			defer input.Close()

			output, err := ioutil.ReadFile(outputFiles[i])
			if err != nil {
				msg := fmt.Sprintf("output test file error: %v", err)
				c.UI.Error(msg)
				return err
			}

			var buf bytes.Buffer
			cmd.Stdin, cmd.Stdout, cmd.Stderr = input, &buf, os.Stderr

			result := judge(cmd, output, v, &info)
			_, testFile := path.Split(inputFiles[i])
			c.UI.Output(fmt.Sprintf("%s\t%s", result, testFile))
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
	-validater, -V       テストの一致方法を指定します (デフォルト diff validater)


`
	return strings.TrimSpace(helpText)
}

func compile(cmds string, lCmd *LangCmd, tmpDir string) error {
	var b bytes.Buffer

	t := template.Must(template.New("compile").Parse(cmds))
	if err := t.Execute(&b, lCmd); err != nil {
		return fmt.Errorf("executing template: %v", err)
	}

	compileCom := strings.Split(b.String(), " ")

	cmd := exec.Command(compileCom[0], compileCom[1:]...)
	cmd.Dir = tmpDir
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", CE, lCmd.File)
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
			return RE
		}

		if !v.Validate(cmd.Stdout.(*bytes.Buffer).Bytes(), expected) {
			return fmt.Sprintf("%s: %d ms", WA, t)
		}

		return fmt.Sprintf("%s: %d ms", AC, t)
	case <-time.After(time.Duration(i.Time) * time.Second):
		return TLE
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
	return "", fmt.Errorf("missing .class file")
}
