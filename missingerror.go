package missingerror

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "missingerror finds errors witch are not returned from function"

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "missingerror",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

var wrappers string // -wrappers flag

func init() {
	Analyzer.Flags.StringVar(&wrappers, "wrappers", "fmt.Errorf",
		"comma-separated list of functions which wrap error")
}

func run(pass *analysis.Pass) (any, error) {
	var wrapperTypes []types.Type
	for _, ws := range strings.Split(wrappers, ",") {
		split := strings.Split(ws, ".")
		if len(split) != 2 {
			return nil, errors.New("invalid flag. wrapper function must be <package>.<name> format")
		}
		wrapperTypes = append(wrapperTypes, analysisutil.TypeOf(pass, split[0], split[1]))
	}

	namedReturnErrors := map[token.Pos]bool{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Type.Results == nil {
					// skip when function returns nothing
					continue
				}
				for _, result := range decl.Type.Results.List {
					for _, name := range result.Names {
						obj := pass.TypesInfo.ObjectOf(name)
						if analysisutil.ImplementsError(obj.Type()) {
							namedReturnErrors[obj.Pos()] = true
						}
					}
				}
			}
		}
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
		(*ast.ReturnStmt)(nil),
	}

	handlingErrors := map[token.Pos]*ast.Ident{}
	var missingErrors []*ast.Ident
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.AssignStmt:
			// 左辺が複数の場合に、右辺にエラーがあるケースで複数回チェックしないようにする
			missing := false
			for i, lh := range n.Lhs {
				if missing {
					break
				}
				switch lh := lh.(type) {
				case *ast.Ident:
					// 変数にerrorインタフェースを実装した型の値を代入する場合、
					// その変数をハンドリング中のエラーとして保持しておく
					// ただし、同一個所で宣言された変数への代入の場合、
					// ハンドリングする前に値が上書きれたとみなして、
					// 既存の値はハンドルされないエラーとして保持する
					obj := pass.TypesInfo.ObjectOf(lh)
					if obj == nil {
						typ := getAssignedType(pass, n, i)
						if analysisutil.ImplementsError(typ) {
							missingErrors = append(missingErrors, lh)
						}
						continue
					}
					ok := analysisutil.ImplementsError(obj.Type())
					if !ok {
						continue
					}
					if defObj := pass.TypesInfo.Defs[lh]; defObj != nil {
						handlingErrors[defObj.Pos()] = lh
					} else if useObj := pass.TypesInfo.Uses[lh]; useObj != nil {
						defPos := useObj.Pos()
						if ident, exist := handlingErrors[defPos]; exist {
							missingErrors = append(missingErrors, ident)
							missing = true
						}
						handlingErrors[defPos] = lh
					}
				}
			}
		case *ast.ReturnStmt:
			for _, result := range n.Results {
				switch result := result.(type) {
				case *ast.Ident:
					obj := pass.TypesInfo.ObjectOf(result)
					ok := analysisutil.ImplementsError(obj.Type())
					if !ok {
						continue
					}
					useObj := pass.TypesInfo.Uses[result]
					delete(handlingErrors, useObj.Pos())
				case *ast.CallExpr:
					if wrappedErrorObj := wrappedError(pass, result, wrapperTypes); wrappedErrorObj != nil {
						delete(handlingErrors, wrappedErrorObj.Pos())
					}
				}
			}
		}
	})
	// maybe sort these properly
	for defPos, e := range handlingErrors {
		if returned := namedReturnErrors[defPos]; returned {
			// returned as named return value
			continue
		}
		missingErrors = append(missingErrors, e)
	}
	sort.Slice(missingErrors, func(i, j int) bool {
		return missingErrors[i].Pos() < missingErrors[j].Pos()
	})
	for _, e := range missingErrors {
		pass.Reportf(e.Pos(), "error wasn't returned")
	}

	return nil, nil
}

func getAssignedType(pass *analysis.Pass, n *ast.AssignStmt, index int) types.Type {
	switch rh := n.Rhs[0].(type) {
	case *ast.CallExpr:
		if tv, ok := pass.TypesInfo.Types[rh.Fun]; ok {
			switch typ := tv.Type.(type) {
			case *types.Signature:
				results := typ.Results()
				return results.At(index).Type()
			}
		}
	}
	return nil
}

func wrappedError(pass *analysis.Pass, result *ast.CallExpr, wrapperTypes []types.Type) types.Object {
	switch f := result.Fun.(type) {
	case *ast.SelectorExpr:
		typ := pass.TypesInfo.TypeOf(f)
		if isWrapperFunction(typ, wrapperTypes) {
			for _, at := range result.Args {
				switch at := at.(type) {
				case *ast.Ident:
					obj := pass.TypesInfo.ObjectOf(at)
					if analysisutil.ImplementsError(obj.Type()) {
						return obj
					}
				}
			}
		}
	}
	return nil
}

func isWrapperFunction(typ types.Type, wrappers []types.Type) bool {
	for _, w := range wrappers {
		if types.Identical(typ, w) {
			return true
		}
	}
	return false
}
