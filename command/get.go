package command

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
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
	No    string
	Name  string
	Level int
	Time  int
	Mem   int
}

const config = "~/.goyuki"
const infoFile = "info.json"
const (
	DPerm = 0755
	FPerm = 0644
)

// Run get test case
func (c *GetCommand) Run(args []string) int {
	if len(args) < 1 {
		msg := fmt.Sprintf("Invalid arguments: %s", strings.Join(args, " "))
		c.UI.Error(msg)
		return 1
	}

	cookie, err := readCookie(config)
	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	if _, err := os.Stat(fmt.Sprint(num)); err == nil {
		c.UI.Error(fmt.Sprintf("Cannot create directory %d: file exists", num))
		return 1
	}

	b, i, err := download(num, cookie)
	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	if err := save(b, i, num); err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	return 0
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

func readCookie(config string) (string, error) {
	cookie := ""

	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GOYUKI") {
			continue
		}
		cookie = strings.Split(v, "=")[1]
		break
	}

	if cookie == "" {
		return "", errors.New("$GOYUKI not set")
	}

	return cookie, nil
}

func download(num int, cookie string) ([]byte, *Info, error) {
	const baseURL = "http://yukicoder.me/problems"
	uri := strings.Join([]string{baseURL, "no", fmt.Sprint(num)}, "/")

	res, err := goreq.Request{
		Uri:          uri,
		MaxRedirects: 1,
	}.Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed problem request : %v", err)
	}
	defer res.Body.Close()

	i, err := parse(res.Body)
	if err != nil {
		return nil, nil, err
	}

	testCaseURI := strings.Join([]string{baseURL, i.No, "testcase.zip"}, "/")
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
		return nil, nil, fmt.Errorf("failed testcase request : %v", err)
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

func save(buf []byte, i *Info, num int) error {
	baseDir := fmt.Sprint(num)
	if err := os.Mkdir(baseDir, DPerm); err != nil {
		return err
	}

	b, err := json.Marshal(*i)
	if err != nil {
		return err
	}

	infoName := strings.Join([]string{baseDir, infoFile}, "/")
	if err := ioutil.WriteFile(infoName, b, FPerm); err != nil {
		return err
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
		return nil, fmt.Errorf("problem parse error : %v", err)
	}

	content := doc.Find("div#content")
	p := content.Find("p")

	i.No, _ = content.Attr("data-problem-id")
	i.Name = content.Find("h3").Text()
	i.Level = p.First().Find("i.fa-star").Size()
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
