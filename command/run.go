package command

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Result is test result
type Result struct {
	info        *Info
	date        time.Time
	compileTime time.Duration
	lang        string
	codeLength  int
}

func (r *Result) String() string {
	s := ""
	switch r.info.JudgeType {
	case Normal:
		s = "Normal"
	case Special:
		s = "Special"
	case Reactive:
		s = "Reactive"
	}
	strs := make([]string, 6)
	strs[0] = fmt.Sprintf("\n問題:\t\t%s", r.info.Name)
	strs[1] = fmt.Sprintf("テスト日時:\t%s", r.date.Format(time.RFC1123))
	strs[2] = fmt.Sprintf("言語:\t\t%s", r.lang)
	strs[3] = fmt.Sprintf("コンパイル時間:\t%d ms", r.compileTime.Nanoseconds()/1000000)
	strs[4] = fmt.Sprintf("コード長:\t%d byte", r.codeLength)
	strs[5] = fmt.Sprintf("ジャッジタイプ:\t%s\n", s)
	return strings.Join(strs, "\n")
}

// Run run the test
func (c *RunCommand) Run(args []string) int {
	var (
		langFlag      string
		validaterFlag string
		verboseFlag   bool
		roundFlag     int
	)

	flags := c.Meta.NewFlagSet("run", c.Help())
	flags.StringVar(&langFlag, "l", "", "Specify Language")
	flags.StringVar(&langFlag, "language", "", "Specify Language")
	flags.StringVar(&validaterFlag, "V", "", "Specify Validater")
	flags.StringVar(&validaterFlag, "validater", "", "Specify Validater")
	flags.BoolVar(&verboseFlag, "vb", false, "increase amount of output")
	flags.BoolVar(&verboseFlag, "verbose", false, "increase amount of output")
	flags.IntVar(&roundFlag, "p", 0, "Rounded to the decimal point p digits")
	flags.IntVar(&roundFlag, "place", 0, "Rounded to the decimal point place digits")

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

	if langFlag == "" {
		langFlag = strings.Replace(path.Ext(args[1]), ".", "", -1)
	}
	lang, ok := Lang[langFlag]
	if !ok {
		msg := fmt.Sprintf("Invalid language: %s", langFlag)
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	if roundFlag < 0 || roundFlag > 15 {
		msg := fmt.Sprintf("Invalid round: %d", roundFlag)
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	if validaterFlag == "float" {
		Validaters["float"] = &FloatValidater{Place: roundFlag}
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

	w, e := ioutil.Discard, ioutil.Discard
	if verboseFlag {
		w, e = os.Stdout, os.Stderr
	}

	code, result, clearFunc, err := NewCode(args[1], lang, &info, w, e)
	if err != nil {
		c.UI.Output(err.Error())
		return ExitCodeFailed
	}
	c.UI.Output(result.String())
	defer clearFunc()

	var rCode *Code
	if info.JudgeType > 0 {
		rCode, clearFunc, err = NewReactiveCode(&info, args[0], w, e)
		if err != nil {
			c.UI.Error(err.Error())
			return ExitCodeFailed
		}
		defer clearFunc()
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
				return fmt.Errorf("input test file error: %v", err)
			}
			defer input.Close()

			output, err := ioutil.ReadFile(outputFiles[i])
			if err != nil {
				return fmt.Errorf("output test file error: %v", err)
			}

			var result string
			if info.JudgeType > 0 {
				result, err = rCode.Reactive(code, inputFiles[i], outputFiles[i], input, w, e)
			} else {
				var buf bytes.Buffer
				result, err = code.Run(v, output, input, &buf, e)
			}
			if err != nil {
				return err
			}

			_, testFile := path.Split(inputFiles[i])
			c.UI.Output(fmt.Sprintf("%s\t%s", result, testFile))
			return nil
		}()
		if err != nil {
			c.UI.Error(err.Error())
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
	-place=n, -p			小数点以下n桁に数値を丸める (float validater時のみ) (0<=n<=15)


`
	return strings.TrimSpace(helpText)
}
