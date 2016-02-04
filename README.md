# goyuki
yukicoderのテストケースをダウンロードしローカルでテストを実行するツール


## Usage
### テストを実行する
コンパイル後、テストを実行する
```bash
$ goyuki run problem_no source_file
```
```bash
-language, -l       実行する言語を指定します
-validater, -V       テストの一致方法を指定します
```

### テストケースを取得する
テストケースを取得する
```bash
$ goyuki get problem_no
```

####問題No.
[No.1 道のショートカット](http://yukicoder.me/problems/17)のテストを実行したい場合は(実行ファイルをmain.goとした場合)
```bash
$ goyuki run 1 main.go
```

テストケースを取得する場合は
```bash
$ goyuki get 1
```

###対応ファイル形式
* C++11
* C
* Java
* Perl (デフォルト)
* Perl6
* PHP
* Python2
* Python3 (デフォルト)
* PyPy2
* PyPy3
* Ruby
* Go
* Haskell
* Scala
* Rust
* Scheme
* OCaml
* JavaScript
* Bash
* Text

## Install

`go get`

```bash
$ go get -d github.com/yukirin/goyuki
```

## Contribution

1. Fork ([https://github.com/yukirin/goyuki/fork](https://github.com/yukirin/goyuki/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[yukirin](https://github.com/yukirin)
