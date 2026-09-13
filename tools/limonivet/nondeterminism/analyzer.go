// Package nondeterminism reports clock and randomness reads inside a Limoni
// model's Init, Update and View.
//
// A session recording replays a model by feeding its Update the recorded
// messages and comparing what View renders. That only works if Update and View
// depend on nothing but the model and the message. A call to time.Now or
// math/rand inside them reads something the recording never captured, so the
// replay renders something else and reports a divergence that is not a bug in
// the model's logic — or worse, hides one.
//
// The fix is to receive the value as a message: limoni.NowCmd for the clock, a
// command returning a random value for randomness. Calls inside a command are
// therefore not reported — a function literal of type func(context.Context) T
// returned from Update runs outside the replay, and its result is recorded.
package nondeterminism

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the limonivet nondeterminism check.
var Analyzer = &analysis.Analyzer{
	Name:     "limoninondeterminism",
	Doc:      "report clock and randomness reads in Limoni model Init, Update and View, which break session replay",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// sources maps a package to the functions in it that read nondeterministic
// state. A nil set means every function and method in the package.
var sources = map[string]map[string]bool{
	"time":         {"Now": true, "Since": true, "Until": true},
	"math/rand":    nil,
	"math/rand/v2": nil,
	"crypto/rand":  nil,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		decl := n.(*ast.FuncDecl)
		if decl.Recv == nil || decl.Body == nil || !isModelMethod(pass, decl) {
			return
		}
		ast.Inspect(decl.Body, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncLit:
				// A command runs outside the model; what it returns is a message,
				// and messages are what a recording keeps.
				return !isCommand(pass, node)
			case *ast.CallExpr:
				if allowed(pass, node) {
					return true
				}
				if fn := calledFunc(pass, node); fn != nil {
					if name, ok := nondeterministic(fn); ok {
						pass.Reportf(node.Pos(),
							"%s in %s cannot be replayed from a session recording; deliver it as a message instead (limoni.NowCmd for the clock)",
							name, decl.Name.Name)
					}
				}
			}
			return true
		})
	})
	return nil, nil
}

// isModelMethod recognises Init() []Cmd, Update(Msg) UpdateResult and
// View(*Frame) by shape, so an unrelated method that happens to be called
// Update is left alone.
func isModelMethod(pass *analysis.Pass, decl *ast.FuncDecl) bool {
	obj, ok := pass.TypesInfo.Defs[decl.Name].(*types.Func)
	if !ok {
		return false
	}
	sig := obj.Type().(*types.Signature)
	params, results := sig.Params().Len(), sig.Results().Len()
	switch decl.Name.Name {
	case "Init":
		return params == 0 && results == 1
	case "Update":
		return params == 1 && results == 1
	case "View":
		if params != 1 || results != 0 {
			return false
		}
		ptr, ok := sig.Params().At(0).Type().(*types.Pointer)
		if !ok {
			return false
		}
		named, ok := ptr.Elem().(*types.Named)
		return ok && named.Obj().Name() == "Frame"
	}
	return false
}

// isCommand reports whether a function literal has the shape of a command:
// one context.Context parameter and one result.
func isCommand(pass *analysis.Pass, lit *ast.FuncLit) bool {
	sig, ok := pass.TypesInfo.TypeOf(lit).(*types.Signature)
	if !ok || sig.Params().Len() != 1 || sig.Results().Len() != 1 {
		return false
	}
	named, ok := sig.Params().At(0).Type().(*types.Named)
	return ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "context" && named.Obj().Name() == "Context"
}

// allowed reports whether the call's line carries a limonivet:allow comment,
// for the rare read that is deliberate — a test fixture proving that replay
// catches nondeterminism, say. The comment keeps the exception visible where
// the call is rather than in a config file somewhere else.
func allowed(pass *analysis.Pass, call *ast.CallExpr) bool {
	pos := pass.Fset.Position(call.Pos())
	for _, file := range pass.Files {
		if pass.Fset.Position(file.Pos()).Filename != pos.Filename {
			continue
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				cpos := pass.Fset.Position(comment.Pos())
				if cpos.Line == pos.Line && strings.Contains(comment.Text, "limonivet:allow") {
					return true
				}
			}
		}
	}
	return false
}

func calledFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	var ident *ast.Ident
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		ident = fun
	case *ast.SelectorExpr:
		ident = fun.Sel
	default:
		return nil
	}
	fn, _ := pass.TypesInfo.Uses[ident].(*types.Func)
	return fn
}

func nondeterministic(fn *types.Func) (string, bool) {
	if fn.Pkg() == nil {
		return "", false
	}
	path := fn.Pkg().Path()
	names, ok := sources[path]
	if !ok {
		return "", false
	}
	if names != nil && !names[fn.Name()] {
		return "", false
	}
	return path + "." + fn.Name(), true
}
