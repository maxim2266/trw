/*
Copyright (c) 2019,2020 Maxim Konakov
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
Package trw wraps around various text processing functions from the standard
Go library to allow for functional composition of operations, also minimising
memory allocations.
*/
package trw

import (
	"bytes"
	"regexp"
)

// Rewriter is an opaque type representing a text rewriting operation.
type Rewriter func([]byte, []byte) ([]byte, []byte)

// fn(dest, src) -> (result, spare)

// Do applies the Rewriter to the specified byte slice. The returned result may be
// either the source slice modified in-place, or a new slice.
func (rw Rewriter) Do(src []byte) (result []byte) {
	result, _ = rw(nil, src)
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

// Matcher is a type of a function that, given a byte slice, returns
// a slice holding the index pairs identifying all successive matches,
// or nil if there is no match.
type Matcher = func([]byte) [][]int

// Delete creates a Rewriter that removes all the matches produced by the given Matcher.
func Delete(match Matcher) Rewriter {
	return func(unused, src []byte) ([]byte, []byte) {
		ms := match(src)

		if len(ms) == 0 {
			return src, unused
		}

		// in-place copy, no need to allocate a new slice
		i, j := ms[0][0], ms[0][1]

		for _, m := range ms[1:] {
			i += copy(src[i:], src[j:m[0]])
			j = m[1]
		}

		if j < len(src) {
			i += copy(src[i:], src[j:])
		}

		return src[:i], unused
	}
}

// Replace creates a rewriter that substitutes all the matches produced by the given Matcher
// with the specified string.
func Replace(match Matcher, subst string) Rewriter {
	if len(subst) == 0 {
		return Delete(match)
	}

	return func(dest, src []byte) ([]byte, []byte) {
		ms := match(src)

		// calculate total length of all matches
		size := 0
		overlap := false

		for _, m := range ms {
			n := m[1] - m[0]

			size += n
			overlap = overlap || n < len(subst)
		}

		// quit if no matches found
		if size == 0 {
			return src, dest
		}

		if overlap {
			// reallocate destination slice if necessary
			if size = len(src) - size + len(ms)*len(subst); size > cap(dest) {
				dest = make([]byte, 0, size+size/5) // +20%
			}

			// copy with replacement
			i := 0

			for _, m := range ms {
				dest = append(append(dest, src[i:m[0]]...), []byte(subst)...)
				i = m[1]
			}

			return append(dest, src[i:]...), src
		}

		// in-place copy with replacement
		i, j := ms[0][0], ms[0][1]

		i += copy(src[i:], subst)

		for _, m := range ms[1:] {
			i += copy(src[i:], src[j:m[0]])
			i += copy(src[i:], subst)
			j = m[1]
		}

		if j < len(src) {
			i += copy(src[i:], src[j:])
		}

		return src[:i], dest
	}
}

// Expand creates a Rewriter that applies Regexp.Expand() operation to every match
// of the given regular expression pattern.
func Expand(patt, subst string) Rewriter {
	return ExpandN(patt, subst, -1)
}

// ExpandN creates a Rewriter that applies Regexp.Expand() operation to the first n matches
// of the given regular expression pattern.
func ExpandN(patt, subst string, n int) Rewriter {
	if len(patt) == 0 {
		panic("empty pattern in trw.ExpandN() function")
	}

	return ExpandReN(regexp.MustCompile(patt), subst, n)
}

// ExpandRe creates a Rewriter that applies Regexp.Expand() operation to all matches
// of the given regular expression object.
func ExpandRe(re *regexp.Regexp, subst string) Rewriter {
	return ExpandReN(re, subst, -1)
}

// ExpandReN creates a Rewriter that applies Regexp.Expand() operation to the first n matches
// of the given regular expression object.
func ExpandReN(re *regexp.Regexp, subst string, n int) Rewriter {
	if re == nil {
		panic("nil regular expression object in trw.ExpandReN() function")
	}

	if len(subst) == 0 {
		return Delete(ReN(re, n))
	}

	return func(dest, src []byte) ([]byte, []byte) {
		ms := re.FindAllSubmatchIndex(src, n)

		if len(ms) == 0 { // avoid copying without a match
			return src, dest
		}

		// (speculatively) reallocate destination slice
		if len(src) > cap(dest) {
			dest = make([]byte, 0, len(src)+len(src)/5) // +20%
		}

		// copy with replacement
		i := 0

		for _, m := range ms {
			dest = re.Expand(append(dest, src[i:m[0]]...), []byte(subst), src, m)
			i = m[1]
		}

		return append(dest, src[i:]...), src
	}
}

// Lit creates a Matcher for the given string literal.
func Lit(patt string) Matcher {
	if len(patt) == 0 {
		panic("empty pattern in trw.Lit() function")
	}

	return func(s []byte) (ms [][]int) {
		for b, i := 0, bytes.Index(s, []byte(patt)); i >= 0; i = bytes.Index(s[b:], []byte(patt)) {
			b += i
			ms = append(ms, []int{b, b + len(patt)})
			b += len(patt)
		}

		return
	}
}

// LitN creates a Matcher for the given string literal that matches up to n times.
func LitN(patt string, n int) Matcher {
	if n < 0 {
		return Lit(patt)
	}

	if len(patt) == 0 {
		panic("empty pattern in trw.LitN() function")
	}

	return func(s []byte) (ms [][]int) {
		for n, b := n, 0; n > 0; n-- {
			if i := bytes.Index(s[b:], []byte(patt)); i >= 0 {
				b += i
				ms = append(ms, []int{b, b + len(patt)})
				b += len(patt)
			} else {
				break
			}
		}

		return
	}
}

// Patt creates a Matcher for the given regular expression pattern.
func Patt(patt string) Matcher {
	return PattN(patt, -1)
}

// PattN creates a Matcher for the given regular expression pattern, matching up to n times.
func PattN(patt string, n int) Matcher {
	if len(patt) == 0 {
		panic("empty pattern in trw.PattN() function")
	}

	return ReN(regexp.MustCompile(patt), n)
}

// Re creates a matcher for the given regular expression object.
func Re(re *regexp.Regexp) Matcher {
	return ReN(re, -1)
}

// ReN creates a matcher for the given regular expression object, matching up to n times.
func ReN(re *regexp.Regexp, n int) Matcher {
	if re == nil {
		panic("nil regular expression object in trw.ReN() function")
	}

	return func(s []byte) [][]int {
		return re.FindAllIndex(s, n)
	}
}
