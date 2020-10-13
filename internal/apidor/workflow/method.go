package workflow

import (
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func findUnusedEndpointMethods(globalMethods []string,
	endpointOperations []definition.EndpointDetails) []string {

	usedEndpointMethods := []string{}
	for _, endpointOperationDetails := range endpointOperations {
		usedEndpointMethods = append(usedEndpointMethods, endpointOperationDetails.Method)
	}

	unusedEndpointMethods := []string{}
	for _, globalMethod := range globalMethods {
		used := false
		for _, usedEndpointMethod := range usedEndpointMethods {
			if globalMethod == usedEndpointMethod {
				used = true
				break
			}
		}
		if !used {
			unusedEndpointMethods = append(unusedEndpointMethods, globalMethod)
		}
	}

	return unusedEndpointMethods
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
