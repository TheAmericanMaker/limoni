// Command limonivet reports code that breaks Limoni session replay.
//
//	go run github.com/thebanri/limoni/tools/limonivet ./...
//
// It is a separate module so that golang.org/x/tools stays out of the Limoni
// module's dependency graph.
package main

import (
	"github.com/thebanri/limoni/tools/limonivet/nondeterminism"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(nondeterminism.Analyzer) }
