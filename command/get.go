package command

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/franela/goreq"
)

// GetCommand is a Command that get test case
type GetCommand struct {
	Meta
}

// Info is problem info
type Info struct {
	No        string
	Name      string
	Level     int
	Time      int
	Mem       int
	RLang     string
	JudgeType int
}

// Run get test case
func (c *GetCommand) Run(args []string) int {
	flags := c.Meta.NewFlagSet("get", c.Help())

	if err := flags.Parse(args); err != nil {
		msg := fmt.Sprintf("Invalid option: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return ExitCodeFailed
	}
	args = flags.Args()

	if len(args) < 1 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return ExitCodeFailed
	}

	cookie := os.Getenv("GOYUKI")
	if cookie == "" {
		c.UI.Error("$GOYUKI not set")
		return ExitCodeFailed
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		c.UI.Error(err.Error())
		return ExitCodeFailed
	}

	if _, err := os.Stat(fmt.Sprint(num)); err == nil {
		c.UI.Error(fmt.Sprintf("Cannot create directory %d: file exists", num))
		return ExitCodeFailed
	}

	b, i, err := download(num, cookie)
	if err != nil {
		c.UI.Error(err.Error())
		return ExitCodeFailed
	}

	rb, err := downloadReactive(i, cookie)
	if err != nil {
		c.UI.Error(err.Error())
		return ExitCodeFailed
	}

	if err := save(b, rb, i, num); err != nil {
		c.UI.Error(err.Error())
		return ExitCodeFailed
	}

	return ExitCodeOK
}

// Synopsis is a one-line, short synopsis of the command.
func (c *GetCommand) Synopsis() string {
	return "テストケースを取得する"
}

// Help is a long-form help text
func (c *GetCommand) Help() string {
	helpText := `
problem_noで指定された番号の問題のテストケースを取得し、
カレントディレクトリに展開する

Usage:
	goyuki get problem_no


`
	return strings.TrimSpace(helpText)
}

func download(num int, cookie string) ([]byte, *Info, error) {
	uri := strings.Join([]string{BaseURL, "no", fmt.Sprint(num)}, "/")

	res, err := goreq.Request{
		Uri:          uri,
		MaxRedirects: 1,
	}.Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed problem request: %v", err)
	}

	if res.StatusCode == 404 {
		return nil, nil, fmt.Errorf("the problem does not exist")
	}

	defer res.Body.Close()

	i, err := parse(res.Body)
	if err != nil {
		return nil, nil, err
	}

	testCaseURI := strings.Join([]string{BaseURL, i.No, "testcase.zip"}, "/")
	session := &http.Cookie{
		Name:     "REVEL_SESSION",
		Value:    cookie,
		Path:     "/",
		HttpOnly: true,
	}

	res, err = goreq.Request{
		Uri: testCaseURI,
	}.WithCookie(session).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed testcase request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 403 {
		return nil, nil, fmt.Errorf("please log in to yukicoder")
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1000000))
	if _, err := io.Copy(buf, res.Body); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), i, nil
}

func save(buf, rbuf []byte, i *Info, num int) error {
	baseDir := fmt.Sprint(num)
	if err := os.Mkdir(baseDir, DPerm); err != nil {
		return err
	}

	b, err := json.Marshal(*i)
	if err != nil {
		return err
	}

	infoName := strings.Join([]string{baseDir, InfoFile}, "/")
	if err := ioutil.WriteFile(infoName, b, FPerm); err != nil {
		return err
	}

	if i.JudgeType > 0 {
		codeName := strings.Join([]string{baseDir, ReactiveCode + Ext(i.RLang)}, "/")
		if err := ioutil.WriteFile(codeName, rbuf, FPerm); err != nil {
			return err
		}
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return err
	}

	var rc io.ReadCloser
	for _, f := range zr.File {
		err := func() error {
			rc, err = f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			path := baseDir + "/" + f.Name
			d := filepath.Dir(path)
			if _, err = os.Stat(d); err != nil {
				if err := os.Mkdir(d, DPerm); err != nil {
					return err
				}
			}

			output, err := os.Create(path)
			if err != nil {
				return err
			}
			defer output.Close()

			if _, err := io.Copy(output, rc); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func parse(r io.Reader) (*Info, error) {
	i := &Info{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("problem parse error: %v", err)
	}

	content := doc.Find("div#content")
	p := content.Find("p")

	i.No, _ = content.Attr("data-problem-id")
	i.Name = content.Find("h3").Text()
	i.Level = p.First().Find("i.fa-star").Size()

	if strings.Contains(content.Text(), "スペシャル") {
		i.JudgeType = Special
	}

	infoData := p.First().Text()

	reg, _ := regexp.Compile(`[\d]+`)
	match := reg.FindAllStringSubmatch(infoData, -1)

	tm := [2]int{}
	for i, v := range match[1:3] {
		n, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, err
		}
		tm[i] = n
	}
	i.Time, i.Mem = tm[0], tm[1]

	return i, nil
}

func downloadReactive(i *Info, cookie string) ([]byte, error) {
	uri := strings.Join([]string{BaseURL, i.No, "code"}, "/")
	session := &http.Cookie{
		Name:     "REVEL_SESSION",
		Value:    cookie,
		Path:     "/",
		HttpOnly: true,
	}

	res, err := goreq.Request{
		Uri:          uri,
		MaxRedirects: 1,
	}.WithCookie(session).Do()
	if err != nil {
		return nil, fmt.Errorf("failed problem request: %v", err)
	}
	defer res.Body.Close()

	buf, err := parseReactive(res.Body, i)
	if err != nil {
		return nil, fmt.Errorf("reactive code parse error: %v", err)
	}
	return buf, nil
}

func parseReactive(r io.Reader, i *Info) ([]byte, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("judge code parse error: %v", err)
	}

	isReactive := false
	doc.Find("option").EachWithBreak(func(n int, e *goquery.Selection) bool {
		if _, ok := e.Attr("selected"); !ok {
			return true
		}

		lang := e.Text()
		i.RLang, isReactive = lang[:strings.Index(lang, " ")], true
		return false
	})

	if isReactive {
		if i.JudgeType != Special {
			i.JudgeType = Reactive
		}
	} else {
		i.JudgeType = Normal
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1000000))
	if _, err := buf.WriteString(doc.Find("textarea").Text()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
