/*
Copyright (c) 2019 Maxim Konakov
All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.
3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software without
   specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY
OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE,
EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

/*
Package trw is a wrapper around various text processing functions from the standard
Go library to allow for functional composition of operations, also minimising
memory allocations during text processing.
*/
package trw

import (
	"bytes"
	"io/ioutil"
	"regexp"
)

// Rewriter is an opaque type representing a text rewriting operation.
type Rewriter func([]byte, []byte) ([]byte, []byte)

// fn(dest, src) -> (result, spare)

// Do applies the Rewriter to the specified byte slice.
func (rw Rewriter) Do(src []byte) (result []byte) {
	result, _ = rw(nil, src)
	return
}

// DoFile applies the Rewriter to the content of the specified file.
func (rw Rewriter) DoFile(name string) (result []byte, err error) {
	if result, err = ioutil.ReadFile(name); err == nil {
		result = rw.Do(result)
	}

	return
}

// Seq is a sequential composition of Rewriters.
func Seq(rewriters ...Rewriter) Rewriter {
	switch len(rewriters) {
	case 0:
		panic("empty Rewriter list in trw.Seq() function")
	case 1:
		return rewriters[0]
	default:
		return func(dest, src []byte) ([]byte, []byte) {
			dest, src = src, dest

			for _, fn := range rewriters {
				dest, src = fn(src[:0], dest)
			}

			return dest, src
		}
	}
}

// Matcher is a type of a function that, given a byte slice, returns a pair of indices
// in that slice defining the location of the first match, or (-1, -1) if there is no match.
type Matcher = func([]byte) (begin, end int)

// Delete creates a Rewriter that removes all the matches produced by the given Matcher.
func Delete(match Matcher) Rewriter {
	return func(unused, src []byte) ([]byte, []byte) {
		b, e := match(src)

		if b < 0 {
			return src, unused
		}

		d, s := src[:b], src[e:]

		for b, e = match(s); b >= 0; b, e = match(s) {
			d, s = append(d, s[:b]...), s[e:]
		}

		return append(d, s...), unused
	}
}

// Replace creates a rewriter that substitutes all the matches produced by the given Matcher
// with the specified string.
func Replace(match Matcher, repl string) Rewriter {
	if len(repl) == 0 {
		return Delete(match)
	}

	return func(dest, src []byte) ([]byte, []byte) {
		// first match
		b, e := match(src)

		if b < 0 { // avoid copying without a match
			return src, dest
		}

		// allocate buffer if not yet
		if dest == nil {
			dest = make([]byte, 0, max(len(src), b+len(repl)))
		}

		// process matches
		d, s := concat(dest, src, b, e, repl)

		for b, e = match(s); b >= 0; b, e = match(s) {
			d, s = concat(d, s, b, e, repl)
		}

		return append(d, s...), src
	}
}

// Expand creates a Rewriter that applies Regexp.Expand() operation to every match
// of the given regular expression.
func Expand(regex, subst string) Rewriter {
	re := regexp.MustCompile(regex)

	if len(subst) == 0 {
		return Delete(mapMatcher(re.FindIndex))
	}

	return func(dest, src []byte) ([]byte, []byte) {
		locs := re.FindAllSubmatchIndex(src, -1)

		if len(locs) == 0 {
			return src, dest
		}

		i := 0

		for _, loc := range locs {
			dest = re.Expand(append(dest, src[i:loc[0]]...), []byte(subst), src, loc)
			i = loc[1]
		}

		return append(dest, src[i:]...), src
	}
}

// Lit creates a Matcher for the given string literal.
func Lit(patt string) Matcher {
	if len(patt) == 0 {
		panic("Empty pattern in str.Lit() function")
	}

	return func(s []byte) (int, int) {
		if i := bytes.Index(s, []byte(patt)); i >= 0 {
			return i, i + len(patt)
		}

		return -1, -1
	}
}

// Regex creates a Matcher for the given regular expression.
func Regex(patt string) Matcher {
	return mapMatcher(regexp.MustCompile(patt).FindIndex)
}

// helpers
func concat(d, s []byte, b, e int, repl string) ([]byte, []byte) {
	return append(append(d, s[:b]...), repl...), s[e:]
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func mapMatcher(match func([]byte) []int) func([]byte) (int, int) {
	return func(s []byte) (int, int) {
		if m := match(s); len(m) > 1 {
			return m[0], m[1]
		}

		return -1, -1
	}
}
