package sourceDataUtil

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
)

type ImportSpecInfo struct {
	Value string
	Name  string
	Path  string
}

type CommonInfo struct {
	Name        string
	Path        string
	PackageName string
}

type FuncDeclInfo struct {
	CommonInfo
	ReturnType *ast.FieldList
	Value      *ast.FuncDecl
	IsVoid     bool
}

type CustomTypeChildInfo struct {
	Name  string
	Value []Props
}

type EnumInfo struct {
	Type  string
	Name  string
	Value string
}

type TypeSpecInfo struct {
	CommonInfo
	CustomTypeChild []CustomTypeChildInfo
	IsEnum          bool
	IsArray         bool
	Value           []Props
	ValueCustom     string
	ValueEnumSpec   *[]EnumInfo
}

func (ty *TypeSpecInfo) getExpr(key string, expr ast.Expr) string {
	var result string

	switch filedType := expr.(type) {
	case *ast.Ident:
		result = filedType.Name
	case *ast.SelectorExpr:
		result = filedType.X.(*ast.Ident).Name + "." + filedType.Sel.Name
	case *ast.ArrayType:
		result = "[]" + ty.getExpr(key, filedType.Elt)
	case *ast.StructType:
		nameChild := ty.Name + key

		if len(filedType.Fields.List) > 0 && key != "" {
			childType := ty.extractPropInfo(filedType.Fields.List)
			ty.CustomTypeChild = append(ty.CustomTypeChild, CustomTypeChildInfo{
				Name:  nameChild,
				Value: childType,
			})
			result = nameChild
		} else {
			fmt.Println("lack element in getExpr", reflect.TypeOf(filedType))
		}

	case *ast.StarExpr:
		result = ty.getExpr(key, filedType.X)
	case *ast.ChanType:
		result = ty.getExpr(key, filedType.Value)
	case *ast.InterfaceType:
		result = "interface{}"
	default:
		fmt.Println("lack element in getExpr", reflect.TypeOf(filedType))
		fmt.Println("lack element in getExpr", filedType)
	}

	return result
}

func (ty *TypeSpecInfo) extractPropInfo(modalData []*ast.Field) []Props {
	var modalInfo []Props

	for _, field := range modalData {
		// key := field.Names[0].Name
		//TODO: handle key
		key := ""
		if len(field.Names) > 0 {
			key = field.Names[0].Name
		}
		keyJson := ""
		value := ty.getExpr(key, field.Type)

		tagValue := ""
		if field.Tag != nil {
			regex := regexp.MustCompile("[\"|\\`]")
			tagValue = regex.ReplaceAllString(field.Tag.Value, "")
		}

		tag, isFound := lo.Find(strings.Split(tagValue, " "), func(item string) bool {
			return strings.Contains(item, "json:")
		})
		if isFound {
			jsonKey := strings.Split(tag, ",")[0]
			keyJson = strings.Split(jsonKey, ":")[1]
		} else {
			if field.Tag != nil {
				fmt.Println("tag", ty.Name, key, ty.Path)
			}
		}

		var isRequired bool
		validateList := []string{"required", "page", "page_size", "not null", "json:id"}
		for _, validate := range validateList {
			if strings.Contains(tagValue, validate) {
				isRequired = true
				break
			}
		}

		modalInfo = append(modalInfo, Props{
			Key:      key,
			KeyJson:  keyJson,
			Value:    value,
			Required: isRequired,
		})
	}

	return modalInfo
}

type Props struct {
	Key      string
	KeyJson  string
	Value    string
	Required bool
}

type AllElements struct {
	ImportSpec      map[string]ImportSpecInfo
	TypeSpec        map[string]TypeSpecInfo
	FuncDecl        map[string]FuncDeclInfo
	EnumData        map[string]TypeSpecInfo
	ListStructArray map[string]bool
}

func PrintJson(data interface{}) {
	dataJson, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(dataJson))
}

func GetSourceInfo() (AllElements, map[string]string, string, error) {
	filePath := "main.go"
	elements := AllElements{
		ImportSpec:      make(map[string]ImportSpecInfo),
		TypeSpec:        make(map[string]TypeSpecInfo),
		FuncDecl:        make(map[string]FuncDeclInfo),
		EnumData:        make(map[string]TypeSpecInfo),
		ListStructArray: make(map[string]bool),
	}
	err := processFile(filePath, &elements, "main")
	if err != nil {
		fmt.Println("Error:", err)
		return elements, nil, "", err
	}

	for _, typeSpec := range elements.TypeSpec {
		if typeSpec.IsArray {
			elements.ListStructArray[typeSpec.Name] = true
		}

		if typeSpec.IsEnum {
			elements.EnumData[typeSpec.Name] = typeSpec
		}
	}

	importPath, allStruct := handleGenAllStruct(elements, elements.ListStructArray)

	return elements, importPath, allStruct, nil
}

func handleGenAllStruct(elements AllElements, listStructArray map[string]bool) (map[string]string, string) {
	var allStruct string

	importPath := make(map[string]string)

	//sort elements.TypeSpec to a-z
	var keys []string
	for k := range elements.TypeSpec {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		typeSpec := elements.TypeSpec[key]
		for _, customTypeChild := range typeSpec.CustomTypeChild {
			allStruct += genStruct(TypeSpecInfo{
				CommonInfo: CommonInfo{
					Name:        customTypeChild.Name,
					PackageName: typeSpec.PackageName,
				},
				Value: customTypeChild.Value,
			}, elements, importPath, listStructArray)
		}
		allStruct += genStruct(typeSpec, elements, importPath, listStructArray)
	}

	return importPath, allStruct
}

func genStruct(typeSpec TypeSpecInfo, elements AllElements, importPath map[string]string, listStructArray map[string]bool) string {
	var allComment string
	if typeSpec.IsEnum {
		allComment += fmt.Sprintf("type %s %s\n\n\n", typeSpec.Name, typeSpec.ValueCustom)

		if len(*typeSpec.ValueEnumSpec) > 0 {
			allComment += fmt.Sprintf("const (\n")
			for _, value := range *typeSpec.ValueEnumSpec {
				allComment += fmt.Sprintf("\t%s_ENUM %s = \"%s\"\n", value.Name, typeSpec.Name, value.Value)
			}
			allComment += fmt.Sprintf(")\n\n\n")
		}

		return allComment
	}

	// allComment += lo.If(typeSpec.IsArray, fmt.Sprintf("type %s []struct {\n", typeSpec.Name)).Else(fmt.Sprintf("type %s struct {\n", typeSpec.Name))
	allComment += lo.If(typeSpec.IsArray, fmt.Sprintf("type %s struct {\n", typeSpec.Name)).Else(fmt.Sprintf("type %s struct {\n", typeSpec.Name))

	for _, props := range typeSpec.Value {
		dataValue := props.Value
		valueFormat := strings.Replace(dataValue, "[]", "", 1)
		value := dataValue
		if strings.Contains(valueFormat, ".") {
			valueSplit := strings.Split(valueFormat, ".")
			packageName := valueSplit[0]
			valueName := valueSplit[1]
			valuePath := elements.ImportSpec[packageName]
			if strings.Contains(valuePath.Value, "go-api") {
				value = valueName
			} else {
				importPath[valuePath.Name] = valuePath.Value
			}
		}

		commentProps := "`"
		if props.KeyJson != "" {
			commentProps += fmt.Sprintf("json:\"%s\"", props.KeyJson)
		}
		if props.Required {
			commentProps += fmt.Sprintf(" validate:\"required\"")
		}
		commentProps += "`"

		isArray := listStructArray[value]
		if !strings.Contains(value, "[]") && (isArray || strings.Contains(dataValue, "[]")) {
			value = fmt.Sprintf("[]%s", value)
		}

		allComment += lo.If(len(commentProps) == 2, fmt.Sprintf("\t%s %s\n", props.Key, value)).Else(fmt.Sprintf("\t%s %s %s\n", props.Key, value, commentProps))
	}
	allComment += "}\n\n"

	return allComment
}

// processFile parses a Go source file and collects its elements.
func processFile(filePath string, elements *AllElements, packagePath string) error {
	if _, processed := elements.ImportSpec[filePath]; processed {
		return nil // Already processed this file, skip
	}

	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	processImports(filePath, node.Imports, elements)
	elements.ImportSpec[filePath] = ImportSpecInfo{} // Mark the file as processed

	enumType := make(map[string][]EnumInfo)
	listEnum := []string{}
	// Collect other elements from the file (you can add more logic here)
	ast.Inspect(node, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.ValueSpec:
			if len(t.Values) > 0 {
				if typeIdent, ok := t.Type.(*ast.Ident); ok {
					typeName := typeIdent.Name
					enumType[typeName] = append(enumType[typeName], EnumInfo{
						Type:  typeName,
						Name:  t.Names[0].Name,
						Value: strings.ReplaceAll(t.Values[0].(*ast.BasicLit).Value, "\"", ""),
					})

				}
			}
		case *ast.GenDecl:
			processTypeSpecs(filePath, packagePath, &listEnum, t.Specs, elements)
		case *ast.FuncDecl:
			processFuncDecl(filePath, packagePath, t, elements)
		}

		return true
	})

	if len(listEnum) > 0 {
		for _, enumName := range listEnum {
			*elements.TypeSpec[enumName].ValueEnumSpec = append(*elements.TypeSpec[enumName].ValueEnumSpec, enumType[enumName]...)
		}
	}

	return nil
}

// processImports processes import specifications and their dependencies.
func processImports(filePath string, imports []*ast.ImportSpec, elements *AllElements) {
	for _, importSpec := range imports {
		packagePath := strings.Trim(importSpec.Path.Value, "\"")

		value := strings.ReplaceAll(importSpec.Path.Value, "\"", "")
		nameFormValue := strings.Split(value, "/")
		name := lo.If(importSpec.Name != nil, importSpec.Name.String()).Else(nameFormValue[len(nameFormValue)-1])
		if strings.Contains(packagePath, "go-api") {
			elements.ImportSpec[name] = ImportSpecInfo{
				Name:  name,
				Value: value,
				Path:  filePath,
			}

			pkg, err := build.Import(packagePath, "", build.FindOnly)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			allFilesInFolder, err := processFolderFindAllFile(pkg.Dir)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			for _, file := range allFilesInFolder {
				if err := processFile(file, elements, packagePath); err != nil {
					fmt.Println("Error:", err)
				}
			}
		} else {
			elements.ImportSpec[name] = ImportSpecInfo{
				Name:  name,
				Value: value,
				Path:  filePath,
			}
		}
	}
}

// processTypeSpecs processes type specifications in a GenDecl.
func processTypeSpecs(filePath string, packageName string, listEnum *[]string, specs []ast.Spec, elements *AllElements) {
	for _, spec := range specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			id := s.Name.Name
			if _, ok := (elements.TypeSpec)[id]; ok {
				fmt.Println("Error: duplicate type name", id)
				continue
			}
			typeSpecInfo := TypeSpecInfo{
				CommonInfo: CommonInfo{
					Name:        id,
					Path:        filePath,
					PackageName: packageName,
				},
				ValueEnumSpec: &[]EnumInfo{},
			}
			switch modalData := s.Type.(type) {
			case *ast.StructType:
				modalDataFiled := modalData.Fields.List
				typeSpecInfo.Value = typeSpecInfo.extractPropInfo(modalDataFiled)
			case *ast.ArrayType:
				typeSpecInfo.IsArray = true
				modalDataFiled := modalData.Elt.(*ast.StructType).Fields.List
				typeSpecInfo.Value = typeSpecInfo.extractPropInfo(modalDataFiled)
			case *ast.Ident:
				typeSpecInfo.IsEnum = true
				typeSpecInfo.ValueCustom = modalData.Name
				*listEnum = append(*listEnum, id)
			default:
				fmt.Println("lack element in processTypeSpecs1", reflect.TypeOf(modalData))
				fmt.Println("lack element in processTypeSpecs2", s.Name)
			}
			(elements.TypeSpec)[id] = typeSpecInfo

		default:
			// fmt.Println("lack element in processTypeSpecs3", s, reflect.TypeOf(s))
		}
	}
}

// processFuncDecl processes function declarations.
func processFuncDecl(filePath string, packageName string, decl *ast.FuncDecl, elements *AllElements) {
	id := fmt.Sprintf("%s.%s", packageName, decl.Name.Name)
	funcDecl := FuncDeclInfo{
		IsVoid:     decl.Type.Results == nil,
		Value:      decl,
		ReturnType: decl.Type.Results,
		CommonInfo: CommonInfo{
			Name:        decl.Name.Name,
			Path:        filePath,
			PackageName: packageName,
		},
	}
	elements.FuncDecl[id] = funcDecl
}

func processFolderFindAllFile(folderPath string) ([]string, error) {
	pkg, err := build.ImportDir(folderPath, 0)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var files []string
	for _, file := range pkg.GoFiles {
		files = append(files, pkg.Dir+"/"+file)
	}

	return files, nil
}
