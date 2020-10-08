package workflow

import (
	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

// Flags is a workflow struct that holds command line flags for customising the workflow
type Flags struct {
}

type apiSummary struct {
	baseURI       string
	authDetails   definition.AuthDetails
	globalHeaders map[string]string
	globalMethods []string
}

// Run is a workflow function that orchestrates the API testing
func Run(definition definition.Definition, flags Flags) {

	apiSummary := apiSummary{
		baseURI:       definition.BaseURI,
		authDetails:   definition.AuthDetails,
		globalHeaders: definition.API.GlobalHeaders,
		globalMethods: definition.API.GlobalMethods,
	}

	for endpoint, endpointDetails := range definition.API.Endpoints {
		requestOptions := buildEndpointRequestOptions(apiSummary, endpoint, endpointDetails)
		testNamePrefix := "token"
		testEndpointWithDifferentAuthTokens(requestOptions, apiSummary.authDetails, testNamePrefix)
	}
}

func testEndpointWithDifferentAuthTokens(requestOptions http.RequestOptions,
	authDetails definition.AuthDetails, testNamePrefix string) {

	// high privileged token
	testName := testNamePrefix + "-high"
	logger.TestPrefix(requestOptions.Endpoint, testName)
	highPrivilegedAuthHeaderValue := buildAuthHeaderValue(authDetails.HeaderValuePrefix, authDetails.High)
	requestOptions.Headers = addHeader(requestOptions.Headers, authDetails.HeaderName, highPrivilegedAuthHeaderValue)
	response := buildAndSendRequest(requestOptions)
	result := verifyResponseExpectedOK(response)
	logger.TestResult(result)

	// low privileged token
	testName = testNamePrefix + "-low"
	logger.TestPrefix(requestOptions.Endpoint, testName)
	lowPrivilegedAuthHeaderValue := buildAuthHeaderValue(authDetails.HeaderValuePrefix, authDetails.Low)
	requestOptions.Headers = addHeader(requestOptions.Headers, authDetails.HeaderName, lowPrivilegedAuthHeaderValue)
	response = buildAndSendRequest(requestOptions)
	result = verifyResponseExpectedUnauthorised(response)
	logger.TestResult(result)

	// no token
	testName = testNamePrefix + "-none"
	logger.TestPrefix(requestOptions.Endpoint, testName)
	requestOptions.Headers = removeHeader(requestOptions.Headers, authDetails.HeaderName)
	response = buildAndSendRequest(requestOptions)
	result = verifyResponseExpectedUnauthorised(response)
	logger.TestResult(result)
}
