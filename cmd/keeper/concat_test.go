package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func BenchmarkConcatStrings(b *testing.B) {
	b.Run("add 70", func(b *testing.B) {
		var f = func(s string) {
			fmt.Fprintf(ioutil.Discard, "keeper: "+s+"\n")
		}
		b.Run("10", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f("aaaaaaaaaa")
			}
		})
		b.Run("70", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			}
		})
	})
	b.Run("build 70", func(b *testing.B) {
		var f = func(s string) {
			var sb strings.Builder
			sb.WriteString("keeper: ")
			sb.WriteString(s)
			sb.WriteString("\n")
			fmt.Fprintf(ioutil.Discard, sb.String())
		}
		b.Run("10", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f("aaaaaaaaaa")
			}
		})
		b.Run("70", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				f("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
			}
		})
	})
}
