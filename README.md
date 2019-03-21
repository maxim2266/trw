# trw: Text Re-Writer.

[![GoDoc](https://godoc.org/github.com/maxim2266/trw?status.svg)](https://godoc.org/github.com/maxim2266/trw)
[![Go report](http://goreportcard.com/badge/maxim2266/trw)](http://goreportcard.com/report/maxim2266/trw)
[![License: BSD 3 Clause](https://img.shields.io/badge/License-BSD_3--Clause-yellow.svg)](https://opensource.org/licenses/BSD-3-Clause)

Package `trw` wraps around various text processing functions from the standard
Go library to allow for functional composition of operations, also minimising
memory allocations during text processing.

#### Usage example:
```Go
func Example() {
	src := []byte("<p>Some    _text_zzz</p>")
	res := rewriter.Do(src)

	fmt.Println(string(res))
	// Output:
	// <p>Some <b>text</b></p>
}

var rewriter = trw.Seq(
	trw.Replace(trw.Regex(`[[:space:]]+`), " "),
	trw.Delete(trw.Lit("zzz")),
	trw.Expand(`_([[:alnum:]]+)_`, `<b>${1}</b>`),
)
```
