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

package trw

import (
	"bytes"
	"testing"
)

func TestDeleteLit1(t *testing.T) {
	cases := []struct {
		src, patt, exp string
	}{
		{"abc", "a", "bc"},
		{"abc", "b", "ac"},
		{"abc", "c", "ab"},
		{"abc", "z", "abc"},
		{"aa bb cc aa bb cc", "aa ", "bb cc bb cc"},
		{"aa bb cc aa bb cc", "bb ", "aa cc aa cc"},
		{"aa bb cc aa bb cc", " cc", "aa bb aa bb"},
		{"abcabc", "abc", ""},
	}

	for i, c := range cases {
		if res := Delete(Lit(c.patt)).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestDeletePatt(t *testing.T) {
	cases := []struct {
		src, patt, exp string
	}{
		{"abc", `^a`, "bc"},
		{"abc", `b+`, "ac"},
		{"abc", `c$`, "ab"},
		{"abc", "z", "abc"},
		{"abc", "^abc$", ""},
		{"aa bb cc aa bb cc", `^aa\s+`, "bb cc aa bb cc"},
		{"aa bb cc aa bb cc", `bb\s+`, "aa cc aa cc"},
		{"aa bb cc aa bb cc", `\s+cc$`, "aa bb cc aa bb"},
	}

	for i, c := range cases {
		if res := Delete(Patt(c.patt)).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestDeletePattN(t *testing.T) {
	cases := []struct {
		src, patt, exp string
	}{
		{"aa bb cc aa bb cc aa bb cc", `a+\s+`, "bb cc bb cc aa bb cc"},
		{"aa bb cc aa bb cc aa bb cc", `b+\s+`, "aa cc aa cc aa bb cc"},
		{"aa bb cc aa bb cc aa bb cc", `c+\s+`, "aa bb aa bb aa bb cc"},
		{"aa bb cc", `a+\s+`, "bb cc"},
	}

	for i, c := range cases {
		if res := Delete(PattN(c.patt, 2)).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestDeleteLitMulti(t *testing.T) {
	cases := []struct {
		src  string
		patt []string
		exp  string
	}{
		{"abc", []string{"a", "b"}, "c"},
		{"abc", []string{"a", "c"}, "b"},
		{"abc", []string{"b", "c"}, "a"},
		{"abc", []string{"a", "z"}, "bc"},
		{"abc", []string{"a", "b", "c"}, ""},
		{"abc", []string{"x", "y", "z"}, "abc"},
	}

	for i, c := range cases {
		rw := make([]Rewriter, 0, len(c.patt))

		for _, p := range c.patt {
			rw = append(rw, Delete(Lit(p)))
		}

		if res := Seq(rw...).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func BenchmarkDelete(b *testing.B) {
	src := []byte("aa bb cc dd")
	exp := []byte("aa cc dd")
	s := make([]byte, len(src), max(len(src), len(exp)))
	fn := Delete(Lit(" bb")).Do
	ok := true

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N && ok; n++ {
		s = s[:len(src)]
		copy(s, src)
		ok = bytes.Equal(fn(s), exp)
	}

	b.StopTimer()

	if !ok {
		b.Error("Benchmark failed!")
		return
	}
}

func TestReplace1(t *testing.T) {
	cases := []struct {
		src, patt, repl, exp string
	}{
		{"abc", "a", "Z", "Zbc"},
		{"abc", "b", "Z", "aZc"},
		{"abc", "c", "Z", "abZ"},
		{"abc", "z", "Z", "abc"},
		{"aa", "aa", "ZZZ", "ZZZ"},
		{"aa bb cc aa bb cc", "aa", "ZZZ", "ZZZ bb cc ZZZ bb cc"},
		{"aa bb cc aa bb cc", "bb", "ZZZ", "aa ZZZ cc aa ZZZ cc"},
		{"aa bb cc aa bb cc", "cc", "ZZZ", "aa bb ZZZ aa bb ZZZ"},
		{"aa bb cc aa bb cc", "zz", "ZZZ", "aa bb cc aa bb cc"},
	}

	for i, c := range cases {
		if res := Replace(Lit(c.patt), c.repl).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestReplaceN(t *testing.T) {
	cases := []struct {
		src, patt, repl, exp string
	}{
		{"aa bb cc aa bb cc aa bb cc", "aa", "ZZZ", "ZZZ bb cc ZZZ bb cc aa bb cc"},
		{"aa bb cc aa bb cc aa bb cc", "bb", "ZZZ", "aa ZZZ cc aa ZZZ cc aa bb cc"},
		{"aa bb cc aa bb cc aa bb cc", "cc", "ZZZ", "aa bb ZZZ aa bb ZZZ aa bb cc"},
		{"aa bb cc aa bb cc aa bb cc", "zz", "ZZZ", "aa bb cc aa bb cc aa bb cc"},
		{"aa bb cc", "cc", "ZZZ", "aa bb ZZZ"},
	}

	for i, c := range cases {
		if res := Replace(LitN(c.patt, 2), c.repl).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestReplaceMulti(t *testing.T) {
	type Subst struct {
		patt, repl string
	}

	cases := []struct {
		src   string
		subst []Subst
		exp   string
	}{
		{"abc", []Subst{{"a", "X"}, {"b", "Y"}, {"c", "Z"}}, "XYZ"},
		{"abc", []Subst{{"a", ""}, {"b", "Y"}, {"c", "Z"}}, "YZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa", "XXX"}, {"bb", "YYY"}, {"cc", "ZZZ"}}, "XXX YYY ZZZ XXX YYY ZZZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa ", ""}, {"bb", "YYY"}, {"cc", "ZZZ"}}, "YYY ZZZ YYY ZZZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa", "XXX"}, {" bb", ""}, {"cc", "ZZZ"}}, "XXX ZZZ XXX ZZZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa", "XXX"}, {"bb", "YYY"}, {" cc", ""}}, "XXX YYY XXX YYY"},
		{"aa bb cc aa bb cc", []Subst{{"bb", "XXX"}, {"XXX", "Y"}, {"Y", "ZZZ"}}, "aa ZZZ cc aa ZZZ cc"},
	}

	for i, c := range cases {
		rw := make([]Rewriter, 0, len(c.subst))

		for _, s := range c.subst {
			rw = append(rw, Replace(Lit(s.patt), s.repl))
		}

		if res := Seq(rw...).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestReplacePatt(t *testing.T) {
	type Subst struct {
		patt, repl string
	}

	cases := []struct {
		src   string
		subst []Subst
		exp   string
	}{
		{"aa bb cc aa bb cc", []Subst{{"a+", "XXX"}, {"b+", "YYY"}, {"c+", "ZZZ"}}, "XXX YYY ZZZ XXX YYY ZZZ"},
		{"aa bb cc aa bb cc", []Subst{{"a+[[:space:]]+", ""}, {"b+", "YYY"}, {"c+", "ZZZ"}}, "YYY ZZZ YYY ZZZ"},
	}

	for i, c := range cases {
		rw := make([]Rewriter, 0, len(c.subst))

		for _, s := range c.subst {
			rw = append(rw, Replace(Patt(s.patt), s.repl))
		}

		if res := Seq(rw...).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func BenchmarkReplace(b *testing.B) {
	src := []byte("aa bb cc dd")
	exp := []byte("X bb cc dd")
	s := make([]byte, len(src), max(len(src), len(exp)))
	fn := Replace(Lit("aa"), "X").Do
	ok := true

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N && ok; n++ {
		s = s[:len(src)]
		copy(s, src)
		ok = bytes.Equal(fn(s), exp)
	}

	b.StopTimer()

	if !ok {
		b.Error("Benchmark failed!")
		return
	}
}

func TestExpand1(t *testing.T) {
	cases := []struct {
		src, patt, repl, exp string
	}{
		{"abc", "(a)", "Z${1}Z", "ZaZbc"},
		{"abc", "(b)", "Z${1}Z", "aZbZc"},
		{"abc", "c", "Z${0}Z", "abZcZ"},
		{"abc", "z", "Z${0}Z", "abc"},
		{"aa bb cc aa bb cc", "aa", "Z${0}Z", "ZaaZ bb cc ZaaZ bb cc"},
		{"aa bb cc aa bb cc", "bb", "Z${0}Z", "aa ZbbZ cc aa ZbbZ cc"},
		{"aa bb cc aa bb cc", "cc", "Z${0}Z", "aa bb ZccZ aa bb ZccZ"},
		{"aa bb cc aa bb cc", "zz", "Z${0}Z", "aa bb cc aa bb cc"},
		{"aaaaaaa bb cc", "aa", "Z${0}Z", "ZaaZZaaZZaaZa bb cc"},
	}

	for i, c := range cases {
		if res := Expand(c.patt, c.repl).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func TestExpandMulti(t *testing.T) {
	type Subst struct {
		patt, repl string
	}

	cases := []struct {
		src   string
		subst []Subst
		exp   string
	}{
		{"aa bb cc aa bb cc", []Subst{{"aa", "X${0}X"}, {"bb", "Y${0}Y"}, {"cc", "Z${0}Z"}}, "XaaX YbbY ZccZ XaaX YbbY ZccZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa ", ""}, {"bb", "Y${0}Y"}, {"cc", "Z${0}Z"}}, "YbbY ZccZ YbbY ZccZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa", "X${0}X"}, {" bb", ""}, {"cc", "Z${0}Z"}}, "XaaX ZccZ XaaX ZccZ"},
		{"aa bb cc aa bb cc", []Subst{{"aa", "X${0}X"}, {"bb", "Y${0}Y"}, {" cc", ""}}, "XaaX YbbY XaaX YbbY"},
		{"aa bb cc aa bb cc", []Subst{{"bb", "X${0}X"}, {"XbbX", "Y"}, {"Y", "ZZZ"}}, "aa ZZZ cc aa ZZZ cc"},
	}

	for i, c := range cases {
		rw := make([]Rewriter, 0, len(c.subst))

		for _, s := range c.subst {
			rw = append(rw, Expand(s.patt, s.repl))
		}

		if res := Seq(rw...).Do([]byte(c.src)); !bytes.Equal(res, []byte(c.exp)) {
			t.Errorf("[%d] Unexpected result: %q instead of %q", i, string(res), c.exp)
			return
		}
	}
}

func BenchmarkExpand(b *testing.B) {
	src := []byte("aa bb cc dd")
	exp := []byte("aa _bb_ cc dd")
	s := make([]byte, len(src), max(len(src), len(exp)))
	fn := Expand("bb", "_${0}_").Do
	ok := true

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N && ok; n++ {
		s = s[:len(src)]
		copy(s, src)
		ok = bytes.Equal(fn(s), exp)
	}

	b.StopTimer()

	if !ok {
		b.Error("Benchmark failed!")
		return
	}
}

// helper functions
func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
