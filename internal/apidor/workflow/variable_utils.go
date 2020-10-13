package workflow

import (
	"strconv"

	"github.com/bncrypted/apidor/pkg/copy"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
	"github.com/bncrypted/apidor/pkg/variable"
)

func duplicateVars(vars []string) []string {
	duplicatedVars := []string{}
	for _, vr := range vars {
		duplicatedVars = append(duplicatedVars, vr+":1")
		duplicatedVars = append(duplicatedVars, vr+":2")
	}
	return duplicatedVars
}

func duplicateRequestParamsWithVars(requestParams map[string]string,
	varsInRequestParams []string) map[string]string {

	requestParamsWithDuplicates := copy.MapOfStrings(requestParams)
	if len(varsInRequestParams) > 0 {
		for key, value := range requestParams {
			if containsVar(value, varsInRequestParams) {
				requestParamsWithDuplicates[key+":1"] = value + ":1"
				requestParamsWithDuplicates[key+":2"] = value + ":2"
				delete(requestParamsWithDuplicates, key)
			}
		}
	}
	return requestParamsWithDuplicates
}

func duplicateBodyParamsWithVars(bodyParams map[string]interface{},
	varsInBodyParams []string) map[string]interface{} {

	return duplicateMapVars(bodyParams, varsInBodyParams)
}

func duplicateMapVars(mp map[string]interface{}, vars []string) map[string]interface{} {
	mapWithDuplicates := copy.Map(mp)
	if len(vars) > 0 {
		for key, value := range mp {
			switch value.(type) {
			case string:
				if containsVar(value.(string), vars) {
					mapWithDuplicates[key+":1"] = value.(string) + ":1"
					mapWithDuplicates[key+":2"] = value.(string) + ":2"
					delete(mapWithDuplicates, key)
				}
			case []interface{}:
				if arr, isVarDirectChild := duplicateArrayVars(value.([]interface{}),
					vars); isVarDirectChild {
					mapWithDuplicates[key+":1"] = addIndexesToVarsInArrayOfStrings(arr, vars, "1")
					mapWithDuplicates[key+":2"] = addIndexesToVarsInArrayOfStrings(arr, vars, "2")
					delete(mapWithDuplicates, key)
				} else {
					mapWithDuplicates[key] = arr
				}
			case map[string]interface{}:
				mapWithDuplicates[key] = duplicateMapVars(value.(map[string]interface{}), vars)
			}
		}
	}
	return mapWithDuplicates
}

func duplicateArrayVars(arr []interface{}, vars []string) ([]interface{}, bool) {
	arrWithDuplicates := copy.Array(arr)
	isVarDirectChild := false
	if len(vars) > 0 {
		for idx, value := range arr {
			switch value.(type) {
			case string:
				if containsVar(value.(string), vars) {
					isVarDirectChild = true
				}
			case []interface{}:
				arrWithDuplicates[idx], _ = duplicateArrayVars(value.([]interface{}), vars)
			case map[string]interface{}:
				arrWithDuplicates[idx] = duplicateMapVars(value.(map[string]interface{}), vars)
			}
		}
	}
	return arrWithDuplicates, isVarDirectChild
}

func addIndexesToVarsInArrayOfStrings(arr []interface{}, vars []string, index string) []interface{} {
	arrWithIndexes := []interface{}{}
	for _, value := range arr {
		switch value.(type) {
		case string:
			if containsVar(value.(string), vars) {
				arrWithIndexes = append(arrWithIndexes, value.(string)+":"+index)
			} else {
				arrWithIndexes = append(arrWithIndexes, value.(string))
			}
		}
	}
	return arrWithIndexes
}

func containsVar(value string, vars []string) bool {
	for _, vr := range vars {
		if vr == value {
			return true
		}
	}
	return false
}

func getVarsFromDefinition(vr string, vrs map[string]definition.Variables) (definition.Variables, bool) {
	// format: $variable
	if definitionVars, ok := vrs[vr[1:]]; ok {
		return definitionVars, true
	}
	// format: $variable$
	if definitonVars, ok := vrs[vr[1:len(vr)-1]]; ok {
		if vr[len(vr)-1] == '$' {
			return definitonVars, true
		}
	}
	// format: $variable:2
	if definitonVars, ok := vrs[vr[1:len(vr)-2]]; ok {
		if vr[len(vr)-2] == ':' && (vr[len(vr)-1]-'0' >= 0 && vr[len(vr)-1]-'0' <= 9) {
			return definitonVars, true
		}
	}
	// format: $variable:2$
	if definitonVars, ok := vrs[vr[1:len(vr)-3]]; ok {
		if vr[len(vr)-3] == ':' && (vr[len(vr)-2]-'0' >= 0 && vr[len(vr)-2]-'0' <= 9) &&
			vr[len(vr)-3] == ':' {
			return definitonVars, true
		}
	}
	return definition.Variables{}, false
}

func getHighPrivilegedVariableValues(requestOptions http.RequestOptions,
	vars map[string]definition.Variables) []string {

	toString := func(value interface{}) (string, bool) {
		switch value.(type) {
		case string:
			return value.(string), true
		case int:
			return strconv.Itoa(value.(int)), true
		}
		return "", false
	}

	varsInRequest := []string{}
	varsInRequest = append(varsInRequest, variable.FindVarsInString(requestOptions.Endpoint)...)
	varsInRequest = append(varsInRequest, variable.FindVarsInMapOfStrings(requestOptions.RequestParams)...)
	varsInRequest = append(varsInRequest, variable.FindVarsInMap(requestOptions.BodyParams)...)

	highPrivilegedVarValues := []string{}
	for _, varInRequest := range varsInRequest {
		if varValues, ok := getVarsFromDefinition(varInRequest, vars); ok {
			if varValue, ok := toString(varValues.High); ok {
				highPrivilegedVarValues = append(highPrivilegedVarValues, varValue)
			}
		}
	}

	return highPrivilegedVarValues
}
