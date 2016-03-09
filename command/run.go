package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		verboseFlag   bool
	)

	flags := c.Meta.NewFlagSet("run", c.Help())
	flags.StringVar(&langFlag, "l", "", "Specify Language")
	flags.StringVar(&langFlag, "language", "", "Specify Language")
	flags.StringVar(&validaterFlag, "V", "", "Specify Validater")
	flags.StringVar(&validaterFlag, "validater", "", "Specify Validater")
	flags.BoolVar(&verboseFlag, "vb", false, "increase amount of output")
	flags.BoolVar(&verboseFlag, "verbose", false, "increase amount of output")

	if err := flags.Parse(args); err != nil {
		msg := fmt.Sprintf("Invalid option: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return ExitCodeFailed
	}
	args = flags.Args()

	if len(args) < 2 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	if _, err := os.Stat(args[0]); err != nil {
		c.UI.Error("does not exist (No such directory)")
		return ExitCodeFailed
	}

	tmpDir, err := ioutil.TempDir("", "goyuki")
	if err != nil {
		c.UI.Error(fmt.Sprintf("can't create directory: %v", err))
		return ExitCodeFailed
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
		return ExitCodeFailed
	}

	if validaterFlag == "" {
		validaterFlag = "diff"
	}
	v, ok := Validaters[validaterFlag]
	if !ok {
		msg := fmt.Sprintf("Invalid validater: %s", validaterFlag)
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	lCmd := LangCmd{
		File: source,
		Exec: strings.Split(source, ".")[0],
	}

	b, err := ioutil.ReadFile(args[1])
	if err != nil {
		msg := fmt.Sprintf("failed to read source file: %v", err)
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	err = ioutil.WriteFile(tmpDir+"/"+source, b, FPerm)
	if err != nil {
		c.UI.Error(err.Error())
		return ExitCodeFailed
	}

	infoBuf, err := ioutil.ReadFile(args[0] + "/" + "info.json")
	if err != nil {
		c.UI.Error(fmt.Sprintf("failed to read info file: %v", err))
		return ExitCodeFailed
	}

	info := Info{}
	if err := json.Unmarshal(infoBuf, &info); err != nil {
		c.UI.Error(err.Error())
		return ExitCodeFailed
	}

	stdout, stderr := ioutil.Discard, ioutil.Discard
	if verboseFlag {
		stdout, stderr = os.Stdout, os.Stderr
	}

	var rCode *Code
	var clearFunc func()
	if info.Reactive {
		rCode, clearFunc, err = reactiveCode(&info, args[0], stdout, stderr)
		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeFailed
		}
		defer clearFunc()
	}

	result := Result{
		info:       &info,
		date:       time.Now(),
		lang:       lang[2],
		codeLength: len(b),
	}

	code := Code{
		LangCmd: &lCmd,
		Lang:    lang,
		Info:    &info,
		Dir:     tmpDir,
	}

	sTime := time.Now()
	err = code.Compile(os.Stdin, stdout, stderr)
	result.compileTime = time.Now().Sub(sTime)
	c.UI.Output(result.String())
	if err != nil {
		c.UI.Output(err.Error())
		return ExitCodeFailed
	}

	if langFlag == "java" || langFlag == "scala" {
		class, err := classFile(tmpDir, source, langFlag)
		lCmd.Class = class

		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeFailed
		}
	}

	inputFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_in", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("input testcase error: %v", err)
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	outputFiles, err := filepath.Glob(strings.Join([]string{args[0], "test_out", "*"}, "/"))
	if err != nil {
		msg := fmt.Sprintf("output testcase error: %v", err)
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	for i := 0; i < len(inputFiles); i++ {
		err := func() error {
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

			var result string
			if info.Reactive {
				result, err = rCode.Reactive(&code, inputFiles[i], outputFiles[i], input, stdout, stderr)
			} else {
				var buf bytes.Buffer
				result, err = code.Run(v, output, input, &buf, stderr)
			}
			if err != nil {
				c.UI.Error(err.Error())
				return err
			}

			_, testFile := path.Split(inputFiles[i])
			c.UI.Output(fmt.Sprintf("%s\t%s", result, testFile))
			return nil
		}()
		if err != nil {
			return ExitCodeFailed
		}
	}
	return ExitCodeOK
}

// Synopsis is a one-line, short synopsis of the command.
func (c *RunCommand) Synopsis() string {
	return "コンパイル後、テストを実行する"
}

// Help is a long-form help text
func (c *RunCommand) Help() string {
	helpText := `
source_fileをコンパイル後、problem_noで指定された番号の問題のテストを実行する

Usage:
	goyuki run problem_no source_file

Options:
	-language=lang, -l		実行する言語を指定します (デフォルト 拡張子から判別)
	-validater=validater, -V      テストの一致方法を指定します (デフォルト diff validater)
	-verbose, -vb		コンパイル時、実行時の標準出力、標準エラー出力を表示する


`
	return strings.TrimSpace(helpText)
}

func reactiveCode(info *Info, no string, w, e io.Writer) (*Code, func(), error) {
	rTmpDir, err := ioutil.TempDir("", "goyuki")
	if err != nil {
		return nil, nil, fmt.Errorf("can't create directory: %v", err)
	}
	clearFunc := func() {
		os.RemoveAll(rTmpDir)
	}

	var lang []string
	var ext string
	for k, v := range Lang {
		if v[2] == info.RLang {
			ext, lang = k, v
			break
		}
	}
	source := ReactiveCode + "." + ext
	lCmd := LangCmd{
		File: source,
		Exec: strings.Split(source, ".")[0],
	}

	b, err := ioutil.ReadFile(no + "/" + source)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read reactive file: %v", err)
	}

	err = ioutil.WriteFile(rTmpDir+"/"+source, b, FPerm)
	if err != nil {
		return nil, nil, err
	}

	rCode := &Code{
		LangCmd: &lCmd,
		Lang:    lang,
		Info:    info,
		Dir:     rTmpDir,
	}

	if err := rCode.Compile(os.Stdin, w, e); err != nil {
		return nil, nil, fmt.Errorf("reactive code compile error")
	}
	return rCode, clearFunc, nil
}

func classFile(dir, source, langFlag string) (string, error) {
	class := ""

	suffix := strings.Split(source, ".")[0] + ".class"
	if langFlag == "scala" {
		suffix = "$.class"
	}

	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		class = strings.TrimSuffix(p, suffix)
		if len(p) != len(class) {
			if langFlag == "java" {
				class += strings.Split(source, ".")[0]
			}

			class = strings.Replace(class, dir+"/", "", -1)
			class = strings.Replace(class, "/", ".", -1)
			return fmt.Errorf("found class")
		}
		return err
	})

	if err.Error() != "found class" {
		return "", fmt.Errorf("missing .class file")
	}
	return class, nil
}
