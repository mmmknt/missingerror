package missingerror

import (
	"go/ast"
	"go/token"
	"sort"

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

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
		(*ast.ReturnStmt)(nil),
	}

	handlingErrors := map[token.Pos]*ast.Ident{}
	var missingErros []*ast.Ident
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.AssignStmt:
			for _, lh := range n.Lhs {
				switch lh := lh.(type) {
				case *ast.Ident:
					// 変数にerrorインタフェースを実装した型の値を代入する場合、
					// その変数をハンドリング中のエラーとして保持しておく
					// ただし、同一個所で宣言された変数への代入の場合、
					// ハンドリングする前に値が上書きれたとみなして、
					// 既存の値はハンドルされないエラーとして保持する
					obj := pass.TypesInfo.ObjectOf(lh)
					// skip when blank identifier
					if obj == nil {
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
							missingErros = append(missingErros, ident)
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
				}
			}
			/*
				case *ast.Ident:
					// 代入、宣言しているerror型の変数があれば、
					// astのnodeを保持
					// その際に、同じtype.Objectが登録されていたら、
					// 既存のnodeをreturnしていないものとして保持
					// astの解析でreturnしていたら、
					// そこで返しているobjectのnodeを解放
					// 最終的に保持しつづけているnodeがnot return
					fmt.Println(n)
					fmt.Println("ast pos:", n.Pos())
					obj := pass.TypesInfo.ObjectOf(n)
					if obj != nil {
						fmt.Println("TypesInfo.ObjectOf:", n, obj)
					}
					defObj := pass.TypesInfo.Defs[n]
					fmt.Println("Defs:", defObj)
					if defObj != nil {
						fmt.Println("Defs:", defObj.Type())
						fmt.Println("Defs:", defObj.Pos())
					}
					useObj := pass.TypesInfo.Uses[n]
					fmt.Println("Uses:", useObj)
					if useObj != nil {
						fmt.Println("Uses:", useObj.Type())
						fmt.Println("Uses:", useObj.Pos())
					}

			*/
		}
	})
	// maybe sort these properly
	for _, e := range handlingErrors {
		missingErros = append(missingErros, e)
	}
	sort.Slice(missingErros, func(i, j int) bool {
		return missingErros[i].Pos() < missingErros[j].Pos()
	})
	for _, e := range missingErros {
		pass.Reportf(e.Pos(), "error wasn't returned")
	}

	return nil, nil
}
