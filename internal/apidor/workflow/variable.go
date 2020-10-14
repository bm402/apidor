package workflow

import (
	"sort"

	"github.com/bncrypted/apidor/internal/apidor/permutation"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
	"github.com/bncrypted/apidor/pkg/variable"
)

func substituteHighPrivilegedVariables(requestOptions http.RequestOptions,
	vars map[string]definition.Variables) http.RequestOptions {

	requestOptions.Endpoint = substituteHighPrivilegedPathParams(requestOptions.Endpoint, vars)
	requestOptions.RequestParams = substituteHighPrivilegedRequestParams(requestOptions.RequestParams, vars)
	requestOptions.BodyParams = substituteHighPrivilegedBodyParams(requestOptions.BodyParams, vars)

	return requestOptions
}

func substituteMixedPrivilegedVariablePermutations(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables) []http.RequestOptions {

	substitutedEndpoints := substituteAllMixedPrivilegedPathParams(
		baseRequestOptions.Endpoint, vars)
	substitutedRequestParams := substituteAllMixedPrivilegedRequestParams(
		baseRequestOptions.RequestParams, vars)
	substitutedBodyParams := substituteAllMixedPrivilegedBodyParams(
		baseRequestOptions.BodyParams, vars)
	return createAllRequestOptions(baseRequestOptions, substitutedEndpoints,
		substitutedRequestParams, substitutedBodyParams)
}

func substituteAndParameterPolluteRequestParams(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables) []http.RequestOptions {

	// parameter pollution on request params
	substitutedRequestParams := []map[string]string{}
	varsInRequestParams := variable.FindVarsInMapOfStrings(baseRequestOptions.RequestParams)
	duplicatedVarsInRequestParams := duplicateVars(varsInRequestParams)
	permutations := permutation.GetCombinationsOfOppositePrivilege(len(duplicatedVarsInRequestParams))
	duplicatedRequestParams := duplicateRequestParamsWithVars(baseRequestOptions.RequestParams,
		varsInRequestParams)

	for _, permutation := range permutations {
		requestParams := substituteMixedPrivilegeRequestParams(duplicatedRequestParams,
			duplicatedVarsInRequestParams, vars, permutation)
		substitutedRequestParams = append(substitutedRequestParams, requestParams)
	}

	substitutedEndpoints := substituteAllMixedPrivilegedPathParams(
		baseRequestOptions.Endpoint, vars)
	substitutedBodyParams := substituteAllMixedPrivilegedBodyParams(
		baseRequestOptions.BodyParams, vars)
	return createAllRequestOptions(baseRequestOptions, substitutedEndpoints,
		substitutedRequestParams, substitutedBodyParams)
}

func substituteAndParameterPolluteBodyParams(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables) []http.RequestOptions {

	// parameter pollution on body params
	substitutedBodyParams := []map[string]interface{}{}
	varsInBodyParams := variable.FindVarsInMap(baseRequestOptions.BodyParams)
	duplicatedVarsInBodyParams := duplicateVars(varsInBodyParams)
	permutations := permutation.GetCombinationsOfOppositePrivilege(len(duplicatedVarsInBodyParams))
	duplicatedBodyParams := duplicateBodyParamsWithVars(baseRequestOptions.BodyParams,
		varsInBodyParams)

	for _, permutation := range permutations {
		bodyParams := substituteMixedPrivilegeBodyParams(duplicatedBodyParams,
			duplicatedVarsInBodyParams, vars, permutation)
		substitutedBodyParams = append(substitutedBodyParams, bodyParams)
	}

	substitutedEndpoints := substituteAllMixedPrivilegedPathParams(
		baseRequestOptions.Endpoint, vars)
	substitutedRequestParams := substituteAllMixedPrivilegedRequestParams(
		baseRequestOptions.RequestParams, vars)
	return createAllRequestOptions(baseRequestOptions, substitutedEndpoints,
		substitutedRequestParams, substitutedBodyParams)
}

func substituteAndParameterWrapBodyParams(baseRequestOptions http.RequestOptions,
	vars map[string]definition.Variables, varsWrappedInArrays map[string]definition.Variables,
	varsWrappedInMaps map[string]definition.Variables) []http.RequestOptions {

	// parameter wrapping on body params
	substitutedBodyParams := []map[string]interface{}{}
	varsInBodyParams := variable.FindVarsInMap(baseRequestOptions.BodyParams)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInBodyParams))

	for _, permutation := range permutations {
		substitutedBodyParams = append(substitutedBodyParams,
			substituteMixedPrivilegeBodyParams(baseRequestOptions.BodyParams, varsInBodyParams,
				varsWrappedInArrays, permutation))
		substitutedBodyParams = append(substitutedBodyParams,
			substituteMixedPrivilegeBodyParams(baseRequestOptions.BodyParams, varsInBodyParams,
				varsWrappedInMaps, permutation))
	}

	substitutedEndpoints := substituteAllMixedPrivilegedPathParams(
		baseRequestOptions.Endpoint, vars)
	substitutedRequestParams := substituteAllMixedPrivilegedRequestParams(
		baseRequestOptions.RequestParams, vars)
	return createAllRequestOptions(baseRequestOptions, substitutedEndpoints,
		substitutedRequestParams, substitutedBodyParams)
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

func substituteAllMixedPrivilegedPathParams(baseEndpoint string,
	vars map[string]definition.Variables) []string {

	substitutedEndpoints := []string{}
	varsInPath := variable.FindVarsInString(baseEndpoint)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInPath))
	for _, permutation := range permutations {
		endpoint := substituteMixedPrivilegePathParams(baseEndpoint,
			varsInPath, vars, permutation)
		substitutedEndpoints = append(substitutedEndpoints, endpoint)
	}
	return substitutedEndpoints
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

func substituteAllMixedPrivilegedRequestParams(baseRequestParams map[string]string,
	vars map[string]definition.Variables) []map[string]string {

	substitutedRequestParams := []map[string]string{}
	varsInRequestParams := variable.FindVarsInMapOfStrings(baseRequestParams)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInRequestParams))
	for _, permutation := range permutations {
		requestParams := substituteMixedPrivilegeRequestParams(baseRequestParams,
			varsInRequestParams, vars, permutation)
		substitutedRequestParams = append(substitutedRequestParams, requestParams)
	}
	return substitutedRequestParams
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

func substituteAllMixedPrivilegedBodyParams(baseBodyParams map[string]interface{},
	vars map[string]definition.Variables) []map[string]interface{} {

	substitutedBodyParams := []map[string]interface{}{}
	varsInBodyParams := variable.FindVarsInMap(baseBodyParams)
	permutations := permutation.GetAllCombinationsOfHighAndLowPrivilege(len(varsInBodyParams))
	for _, permutation := range permutations {
		bodyParams := substituteMixedPrivilegeBodyParams(baseBodyParams,
			varsInBodyParams, vars, permutation)
		substitutedBodyParams = append(substitutedBodyParams, bodyParams)
	}
	return substitutedBodyParams
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

func createAllRequestOptions(baseRequestOptions http.RequestOptions,
	substitutedEndpoints []string, substitutedRequestParams []map[string]string,
	substitutedBodyParams []map[string]interface{}) []http.RequestOptions {

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
