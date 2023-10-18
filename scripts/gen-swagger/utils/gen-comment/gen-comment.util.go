package genCommentUtil

import (
	"botp-gateway/scripts/gen-swagger/types"
	commonUtil "botp-gateway/scripts/gen-swagger/utils/common"
	convertTypeUtil "botp-gateway/scripts/gen-swagger/utils/convert-type"
	sourceDataUtil "botp-gateway/scripts/gen-swagger/utils/source-data"
	"fmt"
	"sort"

	"go/ast"
	"strings"

	"github.com/samber/lo"
)

type SwaggerComment struct {
	AllElements *sourceDataUtil.AllElements
	ImportPath  map[string]string
}

func (sw *SwaggerComment) GenerateSwaggerComments(groupInfos []map[string]types.GroupInfo, responseTypeName map[string]int) string {
	var result string

	var typeInfos []types.GroupInfo
	for _, groupInfo := range groupInfos {
		for _, route := range groupInfo {
			typeInfos = append(typeInfos, route)
		}
	}

	sort.Slice(typeInfos, func(i, j int) bool {
		return typeInfos[i].Path < typeInfos[j].Path || (typeInfos[i].Path == typeInfos[j].Path && typeInfos[i].Variable < typeInfos[j].Variable)
	})

	for _, route := range typeInfos {
		result += sw.handleStruct(route, responseTypeName)
		result += sw.generateGroupComments(route, responseTypeName)
	}

	return result
}

func formatTypeName(name string, packageName string) string {
	packageName = strings.Replace(packageName, "Service", "", 1)
	result := name
	if name == "Create" || name == "Update" || name == "Delete" || name == "Get" {
		result += packageName
	}
	return result
}

func generateStructFromChildData(sw *SwaggerComment, childData []convertTypeUtil.TypePropsCustomResponse, packageName string, responseTypeName map[string]int) string {
	var result string

	for _, data := range childData {
		dataValue := data.Value
		valueFormat := strings.Replace(dataValue, "[]", "", 1)
		value := dataValue
		if strings.Contains(valueFormat, ".") {
			valueSplit := strings.Split(valueFormat, ".")
			packageName := valueSplit[0]
			valueName := valueSplit[1]
			valuePath := sw.AllElements.ImportSpec[packageName]
			if strings.Contains(valuePath.Value, "go-api") {
				value = valueName
			} else {
				sw.ImportPath[valuePath.Name] = valuePath.Value
			}
		}

		isArray := sw.AllElements.ListStructArray[value]
		if !strings.Contains(value, "[]") && (isArray || strings.Contains(dataValue, "[]")) {
			value = fmt.Sprintf("[]%s", value)
		}

		isOverlap := false
		if totalName, ok := responseTypeName[value]; ok && totalName > 1 {
			isOverlap = true
		}

		result += fmt.Sprintf("\t%s %s\n", commonUtil.UpcaseFirstLetter(data.Key), lo.If(isOverlap, value+packageName).Else(value))
	}

	return result
}

func (sw *SwaggerComment) handleStruct(route types.GroupInfo, responseTypeName map[string]int) string {
	var result string

	var keys []string
	for k := range route.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		routeInfo := route.Children[key]
		if routeInfo.Call == nil {
			continue
		}
		packageName := commonUtil.UpcaseFirstLetter(routeInfo.Call.X.(*ast.Ident).Name)
		for _, response := range routeInfo.Response {
			isOverlap := false
			name := formatTypeName(response.Name, packageName)
			packageName = strings.Replace(packageName, "Service", "", 1)
			if totalName, ok := responseTypeName[name]; ok && totalName > 1 {
				isOverlap = true
			}

			if len(*response.ChildResponse) > 0 {
				for childName, childData := range *response.ChildResponse {
					isOverlap := false
					if totalName, ok := responseTypeName[childName]; ok && totalName > 1 {
						isOverlap = true
					}
					result += fmt.Sprintf("type %s struct {\n", lo.If(isOverlap, childName+packageName).Else(childName))
					result += generateStructFromChildData(sw, childData, packageName, responseTypeName)
					result += "}\n\n\r"
				}
			}

			result += lo.If(isOverlap, fmt.Sprintf("type %s%sResponse struct {\n", name, packageName)).Else(fmt.Sprintf("type %sResponse struct {\n", name))
			result += generateStructFromChildData(sw, response.Data, packageName, responseTypeName)
			result += "}\n\n\r"
		}
	}

	return result
}

func (sw *SwaggerComment) generateGroupComments(route types.GroupInfo, responseTypeName map[string]int) string {
	var result string

	var keys []string
	for k := range route.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		routeInfo := route.Children[key]
		if routeInfo.Call == nil {
			continue
		}

		packageName := commonUtil.UpcaseFirstLetter(routeInfo.Call.X.(*ast.Ident).Name)
		router := route.Path + routeInfo.Path

		result += fmt.Sprintf("// @Tags %s\n", strings.Replace(packageName, "Service", "", 1))
		result += generateRouteSummary(routeInfo)

		allParams := getAllParams(router)
		result += allParams

		result += sw.generateRouteParams(routeInfo, sw.AllElements.TypeSpec)
		result += generateRouteResponses(routeInfo, responseTypeName, packageName)
		result += generateRouteRouter(routeInfo, route.Path, routeInfo.Method)
		result += generateSecurity(routeInfo, packageName)
		result += generateRouteFunction(routeInfo, packageName)
		result += "\n\r"
	}

	result += "\n\r"
	return result
}

func getAllParams(router string) string {
	result := ""
	for _, param := range strings.Split(router, "/") {
		if strings.Contains(param, ":") {
			result += fmt.Sprintf("// @Param %s path string true \"%s\"\n", param[1:], param)
		}
	}
	return result
}

func generateRouteSummary(routeInfo types.ServiceInfo) string {
	packageName := commonUtil.UpcaseFirstLetter(routeInfo.Call.X.(*ast.Ident).Name)
	return fmt.Sprintf("// @Summary %s %s\n", routeInfo.Call.Sel, packageName)
}

func (sw *SwaggerComment) generateRouteParams(routeInfo types.ServiceInfo, typeSpec map[string]sourceDataUtil.TypeSpecInfo) string {
	var result string

	if routeInfo.IsBody {
		result += fmt.Sprintf("// @Param payload body %s true \"payload\"\n", routeInfo.NameBody)
	} else {
		for _, props := range routeInfo.Props {
			if props.Key == "" {
				if value, ok := typeSpec[props.Value]; ok {
					for _, data := range value.Value {
						value := getValue(data, sw)

						key := lo.If(props.KeyJson == "", commonUtil.LowercaseFirstLetter(data.Key)).Else(data.KeyJson)
						required := "false"
						if data.Required {
							required = "true"
						}
						result += fmt.Sprintf("// @Param %s %s %s %s \"%s\"\n", key, "query", value, required, key)
					}
				}
			} else {
				value := getValue(props, sw)

				key := lo.If(props.KeyJson == "", commonUtil.LowercaseFirstLetter(props.Key)).Else(props.KeyJson)
				required := "false"
				if props.Required {
					required = "true"
				}

				if typeSpecValue, ok := typeSpec[value]; !ok {
					result += fmt.Sprintf("// @Param %s %s %s %s \"%s\"\n", key, "query", value, required, key)
				} else {
					enumValue := "Enums("
					for _, data := range *typeSpecValue.ValueEnumSpec {
						enumValue += fmt.Sprintf("%s,", data.Value)
					}
					enumValue = strings.TrimRight(enumValue, ",")
					enumValue += ")"

					result += fmt.Sprintf("// @Param %s %s %s %s \"%s\" %s\n", key, "query", "string", required, key, enumValue)
				}
			}

		}
	}

	return result
}

func getValue(props sourceDataUtil.Props, sw *SwaggerComment) string {
	dataValue := props.Value
	valueFormat := strings.Replace(dataValue, "[]", "", 1)
	value := dataValue
	if strings.Contains(valueFormat, ".") {
		valueSplit := strings.Split(valueFormat, ".")
		packageName := valueSplit[0]
		valueName := valueSplit[1]
		valuePath := sw.AllElements.ImportSpec[packageName]
		if strings.Contains(valuePath.Value, "go-api") {
			value = valueName
		} else {
			sw.ImportPath[valuePath.Name] = valuePath.Value
		}
	}

	isArray := sw.AllElements.ListStructArray[value]
	if !strings.Contains(value, "[]") && (isArray || strings.Contains(dataValue, "[]")) {
		value = fmt.Sprintf("[]%s", value)
	}
	return value
}

func generateRouteResponses(routeInfo types.ServiceInfo, responseTypeName map[string]int, packageName string) string {
	var result string

	for _, response := range routeInfo.Response {
		isOverlap := false
		name := formatTypeName(response.Name, packageName)
		packageName = strings.Replace(packageName, "Service", "", 1)
		if totalName, ok := responseTypeName[name]; ok && totalName > 1 {
			isOverlap = true
		}

		result += lo.If(isOverlap == true, fmt.Sprintf("// @Success %d {object} %sResponse\n", response.Status, name+packageName)).Else(fmt.Sprintf("// @Success %d {object} %sResponse\n", response.Status, name))
	}

	return result
}

func generateRouteRouter(routeInfo types.ServiceInfo, groupPath string, method string) string {
	return fmt.Sprintf("// @Router %s [%s]\n", commonUtil.ConvertColonToBraces(groupPath+routeInfo.Path), method)
}

func generateRouteFunction(routeInfo types.ServiceInfo, packageName string) string {
	return fmt.Sprintf("func %s%s() {}\n", routeInfo.Call.Sel, packageName)
}

func generateSecurity(routeInfo types.ServiceInfo, packageName string) string {
	return fmt.Sprintf("// @Security BearerAuth\n")
}
