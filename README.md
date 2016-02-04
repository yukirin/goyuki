# goyuki
yukicoderのテストケースをダウンロードしローカルでテストを実行するツール


## Usage
### `get` コマンド
#### はじめに、GOYUKI環境変数を設定する
yukicoderにログイン(twitter,github)した状態でブラウザのcookieから REVEL\_SESSIONの値を取り出し、GOYUKI環境変数にその値を設定する
```bash
$ export GOYUKI=12345hogehoge # zshの場合
```
#### テストケースを取得する
```bash
$ goyuki get problem_no
```

<br />
### `run` コマンド
#### テストを実行する
コンパイル後、テストを実行する
```bash
$ goyuki run problem_no source_file
```
#### オプション
```bash
-language, -l       実行する言語を指定します (デフォルト 拡張子から判別)
-validater, -V       テストの一致方法を指定します (デフォルト diff validator)
```
##### 例(pypy2でコンパイル、実行し、dif validaterを使用する場合)
```bash
$ goyuki run -l pypy2 -V diff 314 sample.py
```

#### Validater(-validater オプション名)
##### diff Validater(diff)
テストファイルと実行ファイルの出力が完全一致しているか確認する


<br />
### 問題No.
[No.1 道のショートカット](http://yukicoder.me/problems/17)のテストを実行したい場合(ソースファイルをmain.goとした場合)
```bash
$ goyuki run 1 main.go
```

テストケースを取得する場合は
```bash
$ goyuki get 1
```

<br />
### 対応ファイル形式(-language オプション名)
* C++11 (cpp)
* C (c)
* Java (java)
* Perl (pl) (perlのデフォルト)
* Perl6 (pl6)
* PHP (php)
* Python2 (py2)
* Python3 (py) (pythonのデフォルト)
* PyPy2 (pypy2)
* PyPy3 (pypy3)
* Ruby (rb)
* Go (go)
* Haskell (hs)
* Scala (scala)
* Rust (rs)
* Scheme (scm)
* OCaml (ml)
* JavaScript (js)
* Bash (sh)
* Text (txt)

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
