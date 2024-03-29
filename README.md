# trw: Text Re-Writer.

[![GoDoc](https://godoc.org/github.com/maxim2266/trw?status.svg)](https://godoc.org/github.com/maxim2266/trw)
[![Go Report Card](https://goreportcard.com/badge/github.com/maxim2266/trw)](https://goreportcard.com/report/github.com/maxim2266/trw)
[![License: BSD 3 Clause](https://img.shields.io/badge/License-BSD_3--Clause-yellow.svg)](https://opensource.org/licenses/BSD-3-Clause)

Package `trw` wraps around various text processing functions from the standard
Go library to allow for functional composition of operations, also minimising
memory consumption. [Here](example_test.go) is an example of usage.

### About the package

The package is most useful in situations where a number of text rewriting
operations is to be applied sequentially to a large input byte slice. For this scenario the package provides:
- Functional composition of the existing or user-defined operations that can later be
applied all at once;
- Memory optimisation using various techniques to minimise (re)allocations.
