# goyuki
yukicoderのテストケースをダウンロードしローカルでテストを実行するツール(リアクティブジャッジ、スペシャルジャッジ対応)


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
-language=lang, -l       実行する言語を指定します (デフォルト 拡張子から判別)
-validater=validater, -V       テストの一致方法を指定します (デフォルト diff validator)
-verbose, -vb		コンパイル時、実行時の標準出力、標準エラー出力を表示する
-place=n, -p          小数点以下n桁に数値を丸める (float validater時のみ) (0<=n<=15)
```
##### 例(pypy2でコンパイル、実行し、float validaterで出力を小数点以下4桁に丸める場合)
```bash
$ goyuki run -l pypy2 -V float -p 4 314 sample.py
```

#### Validater(-validater オプション名)
リアクティブジャッジ、スペシャルジャッジの場合は無視されます
##### diff Validater(diff)
テストファイルと実行ファイルの出力が行単位で一致しているか確認する
##### float Validater(float)
テストファイルと実行ファイルの出力をFloat64型の数値へ変換し比較する


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
### 対応言語(-language オプション名)
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
* C# (cs)
* D (d)
* Nim (nim)
* Kotlin (kt)
* Crystal (cr)
* F# (fs)
* Fortran (f90)

## Install

`go get`

```bash
$ go get -d github.com/yukirin/goyuki
```
または [Releases yukirin/goyuki Github](https://github.com/yukirin/goyuki/releases)からバイナリをダウンロード

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
