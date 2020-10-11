package workflow

import (
	externalhttp "net/http"
	"time"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

// Flags is a workflow struct that holds command line flags for customising the workflow
type Flags struct {
	Rate int
}

type apiSummary struct {
	baseURI       string
	authDetails   definition.AuthDetails
	globalHeaders map[string]string
	globalMethods []string
}

type verifier func(*externalhttp.Response) string

// Run is a workflow function that orchestrates the API testing
func Run(definition definition.Definition, flags Flags) {

	apiSummary := apiSummary{
		baseURI:       definition.BaseURI,
		authDetails:   definition.AuthDetails,
		globalHeaders: definition.API.GlobalHeaders,
		globalMethods: definition.API.GlobalMethods,
	}

	millisecondsPerRequest := 1000 / flags.Rate
	minRequestDuration := time.Duration(millisecondsPerRequest) * time.Millisecond

	for endpoint, endpointDetails := range definition.API.Endpoints {
		requestOptions := buildEndpointRequestOptions(apiSummary, endpoint, endpointDetails)
		requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)

		testEndpointWithAllLevelsOfAuthentication(requestOptions, apiSummary.authDetails, "token", minRequestDuration)
	}
}

func testEndpointWithAllLevelsOfAuthentication(requestOptions http.RequestOptions,
	authDetails definition.AuthDetails, testNamePrefix string, minRequestDuration time.Duration) {

	testEndpointWithAuthToken(requestOptions, authDetails, testNamePrefix+"-high",
		authDetails.High, verifyResponseExpectedOK, minRequestDuration)
	testEndpointWithAuthToken(requestOptions, authDetails, testNamePrefix+"-low",
		authDetails.Low, verifyResponseExpectedUnauthorised, minRequestDuration)
	testEndpointWithoutAuthToken(requestOptions, authDetails, testNamePrefix+"-none",
		verifyResponseExpectedUnauthorised, minRequestDuration)
}

func testEndpointWithAuthToken(requestOptions http.RequestOptions, authDetails definition.AuthDetails,
	testName string, token string, verifier verifier, minRequestDuration time.Duration) {

	startTime := time.Now()
	logger.TestPrefix(requestOptions.Endpoint, testName)

	authHeaderValue := buildAuthHeaderValue(authDetails.HeaderValuePrefix, token)
	requestOptions.Headers = addHeader(requestOptions.Headers, authDetails.HeaderName, authHeaderValue)

	response, err := buildAndSendRequest(requestOptions)
	if err != nil {
		logger.Message("Skipping due to error: " + err.Error())
		return
	}

	result := verifier(response)
	logger.TestResult(result)

	durationSinceStartTime := time.Since(startTime)
	time.Sleep(minRequestDuration - durationSinceStartTime)
}

func testEndpointWithoutAuthToken(requestOptions http.RequestOptions, authDetails definition.AuthDetails,
	testName string, verifier verifier, minRequestDuration time.Duration) {

	startTime := time.Now()
	logger.TestPrefix(requestOptions.Endpoint, testName)
	requestOptions.Headers = removeHeader(requestOptions.Headers, authDetails.HeaderName)

	response, err := buildAndSendRequest(requestOptions)
	if err != nil {
		logger.Message("Skipping due to error: " + err.Error())
		return
	}

	result := verifier(response)
	logger.TestResult(result)

	durationSinceStartTime := time.Since(startTime)
	time.Sleep(minRequestDuration - durationSinceStartTime)
}
