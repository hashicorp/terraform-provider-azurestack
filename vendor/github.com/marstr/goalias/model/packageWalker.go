package model

import (
	"go/ast"
	"sort"

	"github.com/marstr/collection"
)

// PackageWalker traverses an `*ast.Package` looking for top-level declarations of Constants, Functions, Types, or Variables.
type PackageWalker struct {
	target *ast.Package
}

// Enumerate traverses package content in order of its files lexographically sorted.
func (pw PackageWalker) Enumerate(cancel <-chan struct{}) collection.Enumerator {
	plunger := newDepthBoundPlunger(2)
	go func() {
		defer plunger.Dispose()
		// to ensure the output is deterministic copy the
		// file names into a string slice and sort it.
		files := make([]string, 0, len(pw.target.Files))
		for f := range pw.target.Files {
			files = append(files, f)
		}
		sort.Strings(files)
		for _, f := range files {
			ast.Walk(plunger, pw.target.Files[f])
		}
	}()

	return plunger.Results()
}

type depthBoundPlunger struct {
	currentDepth uint
	MaxDepth     uint
	results      chan interface{}
	Cancel       <-chan struct{}
}

func newDepthBoundPlunger(maxDepth uint) *depthBoundPlunger {
	return &depthBoundPlunger{
		currentDepth: 0,
		MaxDepth:     maxDepth,
		results:      make(chan interface{}),
	}
}

func (plunger depthBoundPlunger) Results() collection.Enumerator {
	return plunger.results
}

func (plunger depthBoundPlunger) Dispose() {
	close(plunger.results)
}

func (plunger depthBoundPlunger) Visit(node ast.Node) (w ast.Visitor) {
	w = nil
	if plunger.currentDepth > plunger.MaxDepth {
		return
	}

	select {
	case <-plunger.Cancel:
		return
	case plunger.results <- node:
		// Intentionally Left Blank
	}

	w = depthBoundPlunger{
		Cancel:       plunger.Cancel,
		currentDepth: plunger.currentDepth + 1,
		MaxDepth:     plunger.MaxDepth,
		results:      plunger.results,
	}

	return
}
