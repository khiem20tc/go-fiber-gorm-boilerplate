package fileUtil

import (
	"botp-gateway/scripts/gen-swagger/types"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type FileAnalyzer struct {
	File    *ast.File
	Imports map[string]*types.ImportSpec
}

func (fi *FileAnalyzer) Init(packageName string) error {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, packageName, nil, parser.AllErrors)

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fi.File = node
	fi.Imports = make(map[string]*types.ImportSpec)
	err = fi.findAllImportSpec()
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}

func (fi *FileAnalyzer) FindFunction(functionNameToFind string) (*ast.FuncDecl, error) {
	var targetFunction *ast.FuncDecl
	for _, decl := range fi.File.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == functionNameToFind {
				targetFunction = fn
				break
			}
		}
	}

	if targetFunction != nil {
		return targetFunction, nil
	}

	return nil, fmt.Errorf("Not found function %s", functionNameToFind)
}

func (fi *FileAnalyzer) findAllImportSpec() error {
	for _, decl := range fi.File.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					packageName := strings.Trim(importSpec.Path.Value, `"`)
					isPackage := strings.Contains(packageName, "go-api")
					if isPackage != true {
						continue
					}
					len := len(strings.Split(packageName, "/"))
					name := strings.Split(packageName, "/")[len-1]
					if importSpec.Name != nil {
						name = importSpec.Name.Name
					}
					importSpec := &types.ImportSpec{
						PackageName: packageName,
						Name:        name,
					}
					fi.Imports[name] = importSpec
				}
			}
		}
	}

	return nil
}

func FindAllElements(targetFunction *ast.FuncDecl) (types.AllElements, error) {
	var allElements types.AllElements
	ast.Inspect(targetFunction.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.AssignStmt:
			allElements.AssignStmt = append(allElements.AssignStmt, n)
		case *ast.Ident:
			allElements.Ident = append(allElements.Ident, n)
		case *ast.CallExpr:
			allElements.CallExpr = append(allElements.CallExpr, n)
		case *ast.BlockStmt:
			allElements.BlockStmt = append(allElements.BlockStmt, n)
		}
		return true
	})
	return allElements, nil
}

func GetPath(path string) string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return dir + path
}

func WriteFile(pathFile string, data []byte) error {
	_, err := os.Stat(pathFile)

	pathFolder := strings.Replace(pathFile, pathFile[strings.LastIndex(pathFile, "/"):], "", 1)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(pathFolder, 0755)
		if errDir != nil {
			return errDir
		}
	}

	f, err := os.Create(pathFile)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
