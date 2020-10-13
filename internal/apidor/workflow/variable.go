package workflow

import (
	"sort"
	"strconv"

	"github.com/bncrypted/apidor/internal/apidor/permutation"
	"github.com/bncrypted/apidor/pkg/copy"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
	"github.com/bncrypted/apidor/pkg/variable"
)

var permutationsCache map[int]string

func substituteHighPrivilegedVariables(requestOptions http.RequestOptions,
	vars map[string]definition.Variables) http.RequestOptions {

	requestOptions.Endpoint = substituteHighPrivilegedPathParams(requestOptions.Endpoint, vars)
	requestOptions.RequestParams = substituteHighPrivilegedRequestParams(requestOptions.RequestParams, vars)
	requestOptions.BodyParams = substituteHighPrivilegedBodyParams(requestOptions.BodyParams, vars)

	return requestOptions
}

func substituteHighPrivilegedPathParams(endpoint string, vars map[string]definition.Variables) string {
	varsInPath := variable.FindVarsInString(endpoint)
	if len(varsInPath) > 0 {
		varsToSubstitute := make(map[string]interface{})
		for _, varInPath := range varsInPath {
			if varToSubstitute, ok := getVarsFromDefinition(varInPath, vars); ok {
				varsToSubstitute[varInPath] = varToSubstitute.High
			}
		}
		endpoint = variable.SubstituteVarsInString(endpoint, varsToSubstitute)
	}
	return endpoint
}

func substituteHighPrivilegedRequestParams(requestParams map[string]string,
	vars map[string]definition.Variables) map[string]string {

	varsInRequestParams := variable.FindVarsInMapOfStrings(requestParams)
	if len(varsInRequestParams) > 0 {
		varsToSubstitute := make(map[string]interface{})
		for _, varInRequestParams := range varsInRequestParams {
			if varToSubstitute, ok := getVarsFromDefinition(varInRequestParams, vars); ok {
				varsToSubstitute[varInRequestParams] = varToSubstitute.High
			}
		}
		requestParams = variable.SubstituteVarsInMapOfStrings(requestParams, varsToSubstitute)
	}
	return requestParams
}

func substituteHighPrivilegedBodyParams(bodyParams map[string]interface{},
	vars map[string]definition.Variables) map[string]interface{} {

	varsInBodyParams := variable.FindVarsInMap(bodyParams)
	if len(varsInBodyParams) > 0 {
		varsToSubstitute := make(map[string]interface{})
		for _, varInBodyParams := range varsInBodyParams {
			if varToSubstitute, ok := getVarsFromDefinition(varInBodyParams, vars); ok {
				varsToSubstitute[varInBodyParams] = varToSubstitute.High
			}
		}
		bodyParams = variable.SubstituteVarsInMap(bodyParams, varsToSubstitute)
	}
	return bodyParams
}

func substituteAllPrivilegedVariablePermutations(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables) []http.RequestOptions {

	// substitute path params
	substitutedEndpoints := []string{}
	varsInPath := variable.FindVarsInString(baseRequestOptions.Endpoint)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInPath))
	for _, permutation := range permutations {
		endpoint := substituteMixedPrivilegePathParams(baseRequestOptions.Endpoint,
			varsInPath, vars, permutation)
		substitutedEndpoints = append(substitutedEndpoints, endpoint)
	}

	// substitute request params
	substitutedRequestParams := []map[string]string{}
	varsInRequestParams := variable.FindVarsInMapOfStrings(baseRequestOptions.RequestParams)
	permutations = permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInRequestParams))
	for _, permutation := range permutations {
		requestParams := substituteMixedPrivilegeRequestParams(baseRequestOptions.RequestParams,
			varsInRequestParams, vars, permutation)
		substitutedRequestParams = append(substitutedRequestParams, requestParams)
	}

	// substitute body params
	substitutedBodyParams := []map[string]interface{}{}
	varsInBodyParams := variable.FindVarsInMap(baseRequestOptions.BodyParams)
	permutations = permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInBodyParams))
	for _, permutation := range permutations {
		bodyParams := substituteMixedPrivilegeBodyParams(baseRequestOptions.BodyParams,
			varsInBodyParams, vars, permutation)
		substitutedBodyParams = append(substitutedBodyParams, bodyParams)
	}

	// create request options for all combinations
	substitutedRequestOptions := []http.RequestOptions{}
	for _, endpoint := range substitutedEndpoints {
		for _, requestParams := range substitutedRequestParams {
			for _, bodyParams := range substitutedBodyParams {
				requestOptions := baseRequestOptions.DeepCopy()
				requestOptions.Endpoint = endpoint
				requestOptions.RequestParams = requestParams
				requestOptions.BodyParams = bodyParams
				substitutedRequestOptions = append(substitutedRequestOptions, requestOptions)
			}
		}
	}

	return substitutedRequestOptions
}

func substituteMixedPrivilegePathParams(endpoint string, varsInPath []string,
	vars map[string]definition.Variables, privilegePermutation string) string {

	if len(varsInPath) > 0 {
		sort.Strings(varsInPath)
		varsToSubstitute := make(map[string]interface{})
		for varInPathIndex, varInPath := range varsInPath {
			if varToSubstitute, ok := getVarsFromDefinition(varInPath, vars); ok {
				if privilegePermutation[varInPathIndex] == 'h' {
					varsToSubstitute[varInPath] = varToSubstitute.High
				} else {
					varsToSubstitute[varInPath] = varToSubstitute.Low
				}
			}
		}
		endpoint = variable.SubstituteVarsInString(endpoint, varsToSubstitute)
	}
	return endpoint
}

func substituteMixedPrivilegeRequestParams(requestParams map[string]string,
	varsInRequestParams []string, vars map[string]definition.Variables,
	privilegePermutation string) map[string]string {

	if len(varsInRequestParams) > 0 {
		sort.Strings(varsInRequestParams)
		varsToSubstitute := make(map[string]interface{})
		for varInPathIndex, varInRequestParams := range varsInRequestParams {
			if varToSubstitute, ok := getVarsFromDefinition(varInRequestParams, vars); ok {
				if privilegePermutation[varInPathIndex] == 'h' {
					varsToSubstitute[varInRequestParams] = varToSubstitute.High
				} else {
					varsToSubstitute[varInRequestParams] = varToSubstitute.Low
				}
			}
		}
		requestParams = variable.SubstituteVarsInMapOfStrings(requestParams, varsToSubstitute)
	}
	return requestParams
}

func substituteMixedPrivilegeBodyParams(bodyParams map[string]interface{},
	varsInBodyParams []string, vars map[string]definition.Variables,
	privilegePermutation string) map[string]interface{} {

	if len(varsInBodyParams) > 0 {
		sort.Strings(varsInBodyParams)
		varsToSubstitute := make(map[string]interface{})
		for varInPathIndex, varInBodyParams := range varsInBodyParams {
			if varToSubstitute, ok := getVarsFromDefinition(varInBodyParams, vars); ok {
				if privilegePermutation[varInPathIndex] == 'h' {
					varsToSubstitute[varInBodyParams] = varToSubstitute.High
				} else {
					varsToSubstitute[varInBodyParams] = varToSubstitute.Low
				}
			}
		}
		bodyParams = variable.SubstituteVarsInMap(bodyParams, varsToSubstitute)
	}
	return bodyParams
}

func substituteOppositePrivilegedRequestParamPermutations(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables) []http.RequestOptions {

	// substitute path params
	substitutedEndpoints := []string{}
	varsInPath := variable.FindVarsInString(baseRequestOptions.Endpoint)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInPath))
	for _, permutation := range permutations {
		endpoint := substituteMixedPrivilegePathParams(baseRequestOptions.Endpoint,
			varsInPath, vars, permutation)
		substitutedEndpoints = append(substitutedEndpoints, endpoint)
	}

	// substitute request params
	substitutedRequestParams := []map[string]string{}
	varsInRequestParams := variable.FindVarsInMapOfStrings(baseRequestOptions.RequestParams)
	duplicatedVarsInRequestParams := duplicateVars(varsInRequestParams)
	permutations = permutation.GetCombinationsOfOppositePrivilege(len(duplicatedVarsInRequestParams))
	duplicatedRequestParams := duplicateRequestParamsWithVars(baseRequestOptions.RequestParams,
		varsInRequestParams)

	for _, permutation := range permutations {
		requestParams := substituteMixedPrivilegeRequestParams(duplicatedRequestParams,
			duplicatedVarsInRequestParams, vars, permutation)
		substitutedRequestParams = append(substitutedRequestParams, requestParams)
	}

	// substitute body params
	substitutedBodyParams := []map[string]interface{}{}
	varsInBodyParams := variable.FindVarsInMap(baseRequestOptions.BodyParams)
	permutations = permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInBodyParams))
	for _, permutation := range permutations {
		bodyParams := substituteMixedPrivilegeBodyParams(baseRequestOptions.BodyParams,
			varsInBodyParams, vars, permutation)
		substitutedBodyParams = append(substitutedBodyParams, bodyParams)
	}

	// create request options for all combinations
	substitutedRequestOptions := []http.RequestOptions{}
	for _, endpoint := range substitutedEndpoints {
		for _, requestParams := range substitutedRequestParams {
			for _, bodyParams := range substitutedBodyParams {
				requestOptions := baseRequestOptions.DeepCopy()
				requestOptions.Endpoint = endpoint
				requestOptions.RequestParams = requestParams
				requestOptions.BodyParams = bodyParams
				substitutedRequestOptions = append(substitutedRequestOptions, requestOptions)
			}
		}
	}

	return substitutedRequestOptions
}

func substituteOppositePrivilegedBodyParamPermutations(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables) []http.RequestOptions {

	// substitute path params
	substitutedEndpoints := []string{}
	varsInPath := variable.FindVarsInString(baseRequestOptions.Endpoint)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInPath))
	for _, permutation := range permutations {
		endpoint := substituteMixedPrivilegePathParams(baseRequestOptions.Endpoint,
			varsInPath, vars, permutation)
		substitutedEndpoints = append(substitutedEndpoints, endpoint)
	}

	// substitute request params
	substitutedRequestParams := []map[string]string{}
	varsInRequestParams := variable.FindVarsInMapOfStrings(baseRequestOptions.RequestParams)
	permutations = permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInRequestParams))
	for _, permutation := range permutations {
		requestParams := substituteMixedPrivilegeRequestParams(baseRequestOptions.RequestParams,
			varsInRequestParams, vars, permutation)
		substitutedRequestParams = append(substitutedRequestParams, requestParams)
	}

	// substitute body params
	substitutedBodyParams := []map[string]interface{}{}
	varsInBodyParams := variable.FindVarsInMap(baseRequestOptions.BodyParams)
	duplicatedVarsInBodyParams := duplicateVars(varsInBodyParams)
	permutations = permutation.GetCombinationsOfOppositePrivilege(len(duplicatedVarsInBodyParams))
	duplicatedBodyParams := duplicateBodyParamsWithVars(baseRequestOptions.BodyParams,
		varsInBodyParams)

	for _, permutation := range permutations {
		bodyParams := substituteMixedPrivilegeBodyParams(duplicatedBodyParams,
			duplicatedVarsInBodyParams, vars, permutation)
		substitutedBodyParams = append(substitutedBodyParams, bodyParams)
	}

	// create request options for all combinations
	substitutedRequestOptions := []http.RequestOptions{}
	for _, endpoint := range substitutedEndpoints {
		for _, requestParams := range substitutedRequestParams {
			for _, bodyParams := range substitutedBodyParams {
				requestOptions := baseRequestOptions.DeepCopy()
				requestOptions.Endpoint = endpoint
				requestOptions.RequestParams = requestParams
				requestOptions.BodyParams = bodyParams
				substitutedRequestOptions = append(substitutedRequestOptions, requestOptions)
			}
		}
	}

	return substitutedRequestOptions
}

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
