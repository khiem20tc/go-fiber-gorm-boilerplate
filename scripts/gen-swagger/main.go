package main

import (
	"botp-gateway/scripts/gen-swagger/types"
	commonUtil "botp-gateway/scripts/gen-swagger/utils/common"
	convertType "botp-gateway/scripts/gen-swagger/utils/convert-type"
	deftSourceUtil "botp-gateway/scripts/gen-swagger/utils/deft-source"
	fileUtil "botp-gateway/scripts/gen-swagger/utils/file"
	genCommentUtil "botp-gateway/scripts/gen-swagger/utils/gen-comment"
	sourceDataUtil "botp-gateway/scripts/gen-swagger/utils/source-data"
	"fmt"
	"go/ast"
	"go/format"
	"strings"
	"time"

	"github.com/samber/lo"
)

// Start: handle RouterInfo
type RouterInfo struct {
	Data        []types.RouteInfo
	allElements types.AllElements
}

type TypeCustomResponse struct {
	name string
	data []convertType.TypePropsCustomResponse
}

type ServiceGroupInfo struct {
	AllResource sourceDataUtil.AllElements
	AllElements types.AllElements
	AllImport   map[string]*types.ImportSpec
	Routes      map[string]types.GroupInfo
	Path        string
	DeftSources map[string]deftSourceUtil.FileTypeInfo
}

func (se *ServiceGroupInfo) ProcessAssignStatements() {
	assignStmts := se.AllElements.AssignStmt

	for _, assignStmt := range assignStmts {
		if len(assignStmt.Lhs) == 1 && len(assignStmt.Rhs) == 1 {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
				if callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr); ok {
					routeInfo := types.GroupInfo{
						Variable: ident.Name,
						Path:     commonUtil.RemoveQuotes(callExpr.Args[0].(*ast.BasicLit).Value),
						Children: make(map[string]types.ServiceInfo),
					}
					se.Routes[ident.Name] = routeInfo
				}
			}
		}
	}
}

func (se *ServiceGroupInfo) ProcessCallExpressions() {
	for _, callExpr := range se.AllElements.CallExpr {
		if funIdent, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if route, ok := se.Routes[funIdent.X.(*ast.Ident).Name]; ok {
				if len(callExpr.Args) > 0 {
					var routeInfo types.ServiceInfo
					var idSelFn string
					for index, ident := range callExpr.Args {
						if index == 0 {
							continue
						}

						if validate, ok := ident.(*ast.CallExpr); ok {
							name := validate.Fun.(*ast.SelectorExpr).Sel.Name
							if name == "ValidateInput" {
								typeName := validate.Args[0].(*ast.CompositeLit).Type.(*ast.SelectorExpr).Sel.Name
								isBody := validate.Args[1].(*ast.Ident).Name == "true"
								typeSpec := se.AllResource.TypeSpec[typeName]
								if isBody {
									routeInfo.NameBody = typeSpec.Name
									routeInfo.IsBody = true
								} else {
									routeInfo.Props = typeSpec.Value
								}
							}

							if strings.Contains(name, "Authen") {
								routeInfo.Authen = true
							}
						}

						if len(callExpr.Args) == index+1 {
							if selectorExpr, ok := ident.(*ast.SelectorExpr); ok {
								routeInfo.Call = selectorExpr

								xName := selectorExpr.X.(*ast.Ident).Name
								idSelFn = selectorExpr.Sel.Name
								packagePath := se.AllImport[xName].PackageName
								id := fmt.Sprintf("%s.%s", packagePath, idSelFn)

								routeInfo.ResponsePath = se.AllResource.FuncDecl[id].Path
							}
						}
					}

					fun := callExpr.Fun.(*ast.SelectorExpr)
					method := fun.Sel.Name
					path := commonUtil.RemoveQuotes(callExpr.Args[0].(*ast.BasicLit).Value)
					path = lo.If(path == "", "/").Else(path)
					routeInfo.Path = path
					routeInfo.Method = method
					idChild := fmt.Sprintf("%s.%s", idSelFn, method)
					route.Children[idChild] = routeInfo
				}
			}
		}
	}
}

func (se *ServiceGroupInfo) UpdateRouteInfoWithResponseData() {
	for _, route := range se.Routes {
		for id, routeInfo := range route.Children {
			if routeInfo.Call == nil || routeInfo.ResponsePath == "" {
				continue
			}

			var fileRespAnalyzer fileUtil.FileAnalyzer
			fileRespAnalyzer.Init(routeInfo.ResponsePath)

			fnResp, err := fileRespAnalyzer.FindFunction(routeInfo.Call.Sel.Name)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}

			var data []convertType.TypeCustomResponse
			ast.Inspect(fnResp.Body, func(n ast.Node) bool {
				if returnStmt, ok := n.(*ast.ReturnStmt); ok {
					result := returnStmt.Results[0]
					packageName := commonUtil.UpcaseFirstLetter(routeInfo.Call.X.(*ast.Ident).Name)
					switch resultType := result.(type) {
					case *ast.CallExpr:
						err := se.processJSONCallExpression(resultType, fnResp.Name.Name, packageName, &data, routeInfo.ResponsePath)
						if err != nil {
							return false
						}
						// default:
						// 	fmt.Println("-----------------------------")
						// 	fmt.Printf("fn.Name.Name %s: path %s\n", fn.Name.Name, routeInfo.ResponsePath)
						// 	fmt.Println("resultType", resultType)
						// 	fmt.Println("resultType", reflect.TypeOf(resultType))
					}
				}

				return true
			})

			routeInfoNew := route.Children[id]
			routeInfoNew.Response = data
			route.Children[id] = routeInfoNew
		}
	}
}

func (se *ServiceGroupInfo) processJSONCallExpression(callExpr *ast.CallExpr, functionName string, packageName string, data *[]convertType.TypeCustomResponse, path string) error {
	if funIdent, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if funIdent.Sel.Name == "JSON" && len(callExpr.Args) > 0 {
			argExpr := callExpr.Args[0]
			status := callExpr.Fun.(*ast.SelectorExpr).X.(*ast.CallExpr).Args[0].(*ast.SelectorExpr).Sel.Name
			statusCode := commonUtil.StatusMap[status]
			if statusCode >= 200 && statusCode < 300 {
				name := functionName
				pathPackage := getPackagePath(path)
				var deftSource deftSourceUtil.FileTypeInfo
				if deftSource, ok = se.DeftSources[pathPackage]; !ok {
					deftSource = deftSourceUtil.GetDeftSource(pathPackage)
					se.DeftSources[pathPackage] = deftSource
				}

				childResponse := make(map[string][]convertType.TypePropsCustomResponse)
				interfaceData := convertType.TypeCustomResponse{
					Name:          name,
					ChildResponse: &childResponse,
					AllResource:   &se.AllResource,
					FileTypeInfo:  &deftSource,
					PackageName:   &packageName,
				}
				interfaceData.Status = statusCode
				interfaceData.Name = name

				interfaceData.Data = interfaceData.ConvertKeyValueToInfo(argExpr.(*ast.CompositeLit), "")
				*data = append(*data, interfaceData)
			}
		}
	}

	return nil
}

func getPackagePath(path string) string {
	pathPackageSplit := strings.Split(path, "/")
	pathPackageSplit = pathPackageSplit[:len(pathPackageSplit)-1]
	pathPackage := strings.Join(pathPackageSplit, "/")
	return pathPackage
}

func main() {
	timeStart := time.Now()
	var timeEnd time.Time
	var (
		err error
	)
	allResource, allImport, allStruct, err := sourceDataUtil.GetSourceInfo()

	routeInfo, err := handleRouter(allResource)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	var responseTypeName = make(map[string]int)
	var allRoutes []map[string]types.GroupInfo
	for _, route := range routeInfo {
		var routerAnalysis fileUtil.FileAnalyzer
		routerAnalysis.Init(route.Path)
		fnCreateRouter, err := routerAnalysis.FindFunction("CreateRouter")
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}

		allElementFnCreateRouter, err := fileUtil.FindAllElements(fnCreateRouter)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}

		serviceGroupInfo := ServiceGroupInfo{
			AllImport:   routerAnalysis.Imports,
			AllResource: allResource,
			AllElements: allElementFnCreateRouter,
			Routes:      make(map[string]types.GroupInfo),
			Path:        route.Path,
			DeftSources: make(map[string]deftSourceUtil.FileTypeInfo),
		}

		serviceGroupInfo.ProcessAssignStatements()
		serviceGroupInfo.ProcessCallExpressions()
		serviceGroupInfo.UpdateRouteInfoWithResponseData()

		allRoutes = append(allRoutes, serviceGroupInfo.Routes)
	}

	for _, route := range allRoutes {
		for _, routeInfo := range route {
			handleStructDuplicateName(routeInfo, responseTypeName)
		}
	}

	var allComments string
	var allService string
	allComments += "//Code generated by gen-swagger. DO NOT EDIT.\n"
	allComments += "package main\n"

	swaggerComment := genCommentUtil.SwaggerComment{
		AllElements: &allResource,
		ImportPath:  allImport,
	}

	allService += swaggerComment.GenerateSwaggerComments(allRoutes, responseTypeName)

	allComments += "import (\n"
	for name, importSpec := range allImport {
		allComments += fmt.Sprintf("\t%s \"%s\"\n", name, importSpec)
	}
	allComments += ")\n"

	allComments += allStruct
	allComments += allService

	allCommentsWhenFormat, _ := format.Source([]byte(allComments))
	err = fileUtil.WriteFile("./gen-swagger/gen-swagger.go", []byte(allCommentsWhenFormat))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	timeEnd = time.Now()
	fmt.Println("Success generate gen-swagger.go in", timeEnd.Sub(timeStart))
}

func handleStructDuplicateName(route types.GroupInfo, responseTypeName map[string]int) {
	for _, routeInfo := range route.Children {
		if routeInfo.Call == nil {
			continue
		}

		for _, response := range routeInfo.Response {
			responseTypeName[response.Name]++
			for key := range *response.ChildResponse {
				responseTypeName[key]++
			}
		}
	}

}

func handleRouter(allResource sourceDataUtil.AllElements) ([]types.RouteInfo, error) {
	filePath := fileUtil.GetPath("/router/router.go")
	functionNameToFind := "New"

	var routerAnalysis fileUtil.FileAnalyzer
	routerAnalysis.Init(filePath)

	funcNew, err := routerAnalysis.FindFunction(functionNameToFind)
	if err != nil {
		return nil, fmt.Errorf("Error: FindFunction New from router")
	}

	allElementFuncNew, err := fileUtil.FindAllElements(funcNew)
	if err != nil {
		return nil, fmt.Errorf("Error: FindAllElements New from router")
	}

	var dataRouterInfo []types.RouteInfo
	for _, callExpr := range allElementFuncNew.CallExpr {
		if funIdent, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if funIdent.Sel.Name == "CreateRouter" {
				xName := funIdent.X.(*ast.Ident).Name
				selName := funIdent.Sel.Name
				packagePath := routerAnalysis.Imports[xName].PackageName
				id := fmt.Sprintf("%s.%s", packagePath, selName)
				dataRouterInfo = append(dataRouterInfo, types.RouteInfo{
					X:           xName,
					Sel:         funIdent.Sel.Name,
					PackageName: packagePath,
					Path:        allResource.FuncDecl[id].Path,
				})
			}
		}
	}

	return dataRouterInfo, nil
}
