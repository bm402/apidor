package workflow

import (
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

func substituteHighPrivilegedPathParams(endpoint string, vars map[string]definition.Variables) string {
	varsInPath := variable.FindVarsInString(endpoint)
	if len(varsInPath) > 0 {
		varsToSubstitute := make(map[string]interface{})
		for _, varInPath := range varsInPath {
			if varToSubstitute, ok := vars[varInPath[1:]]; ok {
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
			if varToSubstitute, ok := vars[varInRequestParams[1:]]; ok {
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
			if varToSubstitute, ok := vars[varInBodyParams[1:]]; ok {
				varsToSubstitute[varInBodyParams] = varToSubstitute.High
			}
		}
		bodyParams = variable.SubstituteVarsInMap(bodyParams, varsToSubstitute)
	}
	return bodyParams
}
