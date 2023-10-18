package convertTypeUtil

import (
	commonUtil "botp-gateway/scripts/gen-swagger/utils/common"
	deftSourceUtil "botp-gateway/scripts/gen-swagger/utils/deft-source"
	sourceDataUtil "botp-gateway/scripts/gen-swagger/utils/source-data"
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/samber/lo"
)

type TypeCustomResponse struct {
	Name          string
	Status        int
	PackageName   *string
	Data          []TypePropsCustomResponse
	ChildResponse *map[string][]TypePropsCustomResponse
	AllResource   *sourceDataUtil.AllElements
	FileTypeInfo  *deftSourceUtil.FileTypeInfo
}

type TypePropsCustomResponse struct {
	Key   string
	Value string
}

func RemoveSpecialCharacters(str string) string {
	re := regexp.MustCompile(`[-\` + "`" + `~!@#$%^&*()_+<>«»?:"{},./;'[]`)
	return re.ReplaceAllString(str, "")
}

func (tc *TypeCustomResponse) ConvertKeyValueToInfo(compLit *ast.CompositeLit, keyApe string) []TypePropsCustomResponse {
	var result []TypePropsCustomResponse

	for _, elt := range compLit.Elts {
		key := RemoveSpecialCharacters(elt.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value)
		keyForStruct := lo.If(len(key) > 0, keyApe+commonUtil.UpcaseFirstLetter(key)).Else(keyApe)

		result = append(result, TypePropsCustomResponse{
			Key:   key,
			Value: tc.getTypeNameFromExpr(elt.(*ast.KeyValueExpr).Value, keyForStruct),
		})
	}

	return result
}

func (tc *TypeCustomResponse) getTypeNameFromExpr(valueExpr ast.Expr, key string) string {
	var result string
	switch value := valueExpr.(type) {

	case *ast.BasicLit:
		result = strings.ToLower(value.Kind.String())
	case *ast.Ident:
		name := value.Name
		result = tc.getTypeNameForIdent(*value, &name)
	case *ast.ArrayType:
		arrayType := tc.getTypeNameFromExpr(value.Elt, "")
		result = fmt.Sprintf("[]%s", arrayType)
	case *ast.SelectorExpr:
		callExprName := deftSourceUtil.GetNameAst(value)
		model := tc.FileTypeInfo.HiddenFunc[callExprName]
		isArray := strings.Contains(model, "[]")
		if isArray {
			result = fmt.Sprint("[]", getLastPath(model))
		} else {
			result = getLastPath(model)
		}
	case *ast.KeyValueExpr:
		result = tc.getTypeNameFromExpr(value.Value, "")
	case *ast.CompositeLit:
		keyName := ""
		if len(key) > 0 && key != "" {
			keyName = commonUtil.UpcaseFirstLetter(key)
		}
		nameResponseChild := tc.Name + keyName
		// }
		data := value.Type.(*ast.SelectorExpr)
		if data.Sel.Name == "Map" {
			(*tc.ChildResponse)[nameResponseChild] = tc.ConvertKeyValueToInfo(value, commonUtil.UpcaseFirstLetter(key))
			result = nameResponseChild
		} else {
			result = fmt.Sprintf("%s.%s", data.X, data.Sel)
		}
	case *ast.CallExpr:
		callExprName := deftSourceUtil.GetNameAst(value)
		model := tc.FileTypeInfo.HiddenFunc[callExprName]
		isArray := strings.Contains(model, "[]")
		if isArray {
			result = fmt.Sprint("[]", getLastPath(model))
		} else {
			result = getLastPath(model)
		}
	case *ast.IndexExpr:
		if _, ok := value.Index.(*ast.BasicLit); ok {
			switch valueX := value.X.(type) {
			case *ast.Ident:
				return tc.getTypeNameForIdent(*valueX, nil)
			default:
				fmt.Println("valueX not define", reflect.TypeOf(valueX))
			}
		}
	case *ast.MapType:
		result = tc.getTypeNameFromExpr(value.Value, "")
	default:
		fmt.Println("valueExpr", reflect.TypeOf(value))
	}

	return result
}

func (tc *TypeCustomResponse) getTypeNameForIdent(valueIdent ast.Ident, name *string) string {
	var result string
	if valueIdent.Obj == nil {
		if valueIdent.Name == "true" || valueIdent.Name == "false" {
			result = "bool"
		} else {
			result = valueIdent.Name
		}
	} else {
		switch decl := valueIdent.Obj.Decl.(type) {
		case *ast.TypeSpec:
			typeOfIdent := decl.Type
			if identEl, ok := typeOfIdent.(*ast.Ident); ok {
				result = identEl.Name
			} else {
				result = tc.getTypeNameFromExpr(typeOfIdent, "")
			}
		case *ast.ValueSpec:
			if decl.Type == nil {
				name := decl.Names[0].Name
				funcInfo := tc.FileTypeInfo.FuncInfo[tc.Name]
				typeAssign := funcInfo.Defs[name].(string)
				result = getLastPath(typeAssign)
			} else {
				switch typeOfIdent := decl.Type.(type) {
				case *ast.Ident:
					result = typeOfIdent.Name
				case *ast.ArrayType:
					result = fmt.Sprintf("[]%s", tc.getTypeNameFromExpr(typeOfIdent.Elt, ""))
				default:
					result = tc.getTypeNameFromExpr(typeOfIdent, "")
				}
			}
		case *ast.AssignStmt:
			for _, lh := range decl.Lhs {
				if lh.(*ast.Ident).Name == *name {
					nameIdent := lh.(*ast.Ident).Name
					// fmt.Println("nameIdent", decl.)
					funcInfo := tc.FileTypeInfo.FuncInfo[tc.Name]
					typeAssign := funcInfo.Defs[nameIdent].(string)
					result = getLastPath(typeAssign)
				}
			}
		case *ast.CompositeLit:
			compositeLit := decl.Type.(*ast.SelectorExpr)
			result = fmt.Sprintf("%s.%s", compositeLit.X, compositeLit.Sel)
		default:
			fmt.Println("valueIdent.Obj.Decl", reflect.TypeOf(decl))
			fmt.Println("valueIdent.Obj.Decl", decl)
		}
	}

	return result
}

func getLastPath(path string) string {
	pathSplit := strings.Split(path, "/")
	return pathSplit[len(pathSplit)-1]
}
