package deftSourceUtil

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"reflect"

	"golang.org/x/tools/go/packages"
)

type FileTypeInfo struct {
	FuncInfo   map[string]FuncInfo
	HiddenFunc map[string]string
}

type FuncInfo struct {
	Name  string
	Value string
	Defs  map[string]interface{}
}

func GetDeftSource(dir string) FileTypeInfo {
	fset := token.NewFileSet()

	config := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo,
		Fset: fset,
		Dir:  dir,
	}

	pkgs, err := packages.Load(config)
	if err != nil {
		log.Fatal(err)
	}

	filePathInfo := FileTypeInfo{
		HiddenFunc: make(map[string]string),
		FuncInfo:   make(map[string]FuncInfo),
	}

	for _, pkg := range pkgs {
		typeInfo := pkg.TypesInfo
		for _, file := range typeInfo.Defs {
			if file == nil {
				continue
			}
			if fn, ok := file.(*types.Func); ok {

				fnInfo := FuncInfo{
					Name:  fn.Name(),
					Value: fn.Type().String(),
				}
				funcName := fn.Name()
				defs := getDefsForFunction(fn.Scope(), typeInfo)
				fnInfo.Defs = defs

				filePathInfo.FuncInfo[funcName] = fnInfo
			}
		}
		for astValue, typeValue := range typeInfo.Types {
			switch typeName := reflect.TypeOf(astValue).String(); typeName {
			case "*ast.CallExpr":
				resultType := astValue.(*ast.CallExpr).Fun
				filePathInfo.HiddenFunc[GetNameAst(resultType)] = typeValue.Type.String()
			case "*ast.SelectorExpr":
				filePathInfo.HiddenFunc[GetNameAst(astValue)] = typeValue.Type.String()
			default:
				// fmt.Println("file", typeName)
			}
		}
	}

	return filePathInfo
}

func getNameIdent(ident *ast.Ident) string {
	return ident.Name
}

func getNameSelectorExpr(selectorExpr ast.SelectorExpr) string {
	return GetNameAst(selectorExpr.X) + "." + selectorExpr.Sel.Name
}

func getNameCallExpr(callExpr ast.CallExpr) string {
	return GetNameAst(callExpr.Fun)
}

func getNameFuncLit(funcLit ast.FuncLit) string {
	return GetNameAst(funcLit.Type)
}

func GetNameAst(astType ast.Expr) string {
	switch typeName := reflect.TypeOf(astType).String(); typeName {
	case "*ast.Ident":
		return getNameIdent(astType.(*ast.Ident))
	case "*ast.SelectorExpr":
		return getNameSelectorExpr(*astType.(*ast.SelectorExpr))
	case "*ast.CallExpr":
		return getNameCallExpr(*astType.(*ast.CallExpr))
	default:
		return ""
	}
}

func getDefsForFunction(node *types.Scope, typeInfo *types.Info) map[string]interface{} {
	defs := make(map[string]interface{})
	for _, name := range node.Names() {
		defs[name] = node.Lookup(name).Type().String()
	}

	return defs
}
