package workflow

import (
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func findUnusedEndpointMethods(globalMethods []string,
	endpointOperations []definition.EndpointDetails) []string {

	usedMethods := make(map[string]bool)
	for _, endpointOperationDetails := range endpointOperations {
		usedMethods[endpointOperationDetails.Method] = true
	}

	unusedMethods := []string{}
	for _, globalMethod := range globalMethods {
		if usedMethods[globalMethod] != true {
			unusedMethods = append(unusedMethods, globalMethod)
		}
	}

	return unusedMethods
}

func substituteUnusedMethods(baseRequestOptions http.RequestOptions,
	unusedEndpointMethods []string) []http.RequestOptions {

	collectedRequestOptions := []http.RequestOptions{}
	for _, unusedEndpointMethod := range unusedEndpointMethods {
		requestOptions := baseRequestOptions.DeepCopy()
		requestOptions.Method = unusedEndpointMethod
		collectedRequestOptions = append(collectedRequestOptions, requestOptions)
	}
	return collectedRequestOptions
}
