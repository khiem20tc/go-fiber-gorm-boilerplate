package types

import (
	convertTypeUtil "botp-gateway/scripts/gen-swagger/utils/convert-type"
	sourceDataUtil "botp-gateway/scripts/gen-swagger/utils/source-data"
	"go/ast"
)

type GroupInfo struct {
	Variable string
	Path     string
	Children map[string]ServiceInfo
}

type ServiceInfo struct {
	Method       string
	Path         string
	Props        []sourceDataUtil.Props
	NameBody     string
	IsBody       bool
	Call         *ast.SelectorExpr
	Response     []convertTypeUtil.TypeCustomResponse
	ResponsePath string
	Authen       bool
}

// X is a package name
// Sel is a function name
type RouteInfo struct {
	X           string
	Sel         string
	PackageName string
	Path        string
}

type ImportSpec struct {
	Path        []string
	PackageName string
	Name        string
}

type AllElements struct {
	AssignStmt []*ast.AssignStmt
	Ident      []*ast.Ident
	CallExpr   []*ast.CallExpr
	BlockStmt  []*ast.BlockStmt
	ImportSpec []*ImportSpec
}
