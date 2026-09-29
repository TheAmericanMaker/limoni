package main

import (
	"strings"
	"testing"
)

// BenchmarkReflowFullHistory reflows a full history of 10,000 lines, the
// worst a drag-resize meets on every size it passes through.
func BenchmarkReflowFullHistory(b *testing.B) {
	var out strings.Builder
	for i := 0; i < 10100; i++ {
		out.WriteString(strings.Repeat("word ", 20) + "\r\n")
	}
	s := reflowSession(120, 40, out.String())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			s.resizeTo(90, 40)
		} else {
			s.resizeTo(120, 40)
		}
	}
}
