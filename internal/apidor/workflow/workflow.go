package workflow

import (
	externalhttp "net/http"
	"strconv"
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

type verifier func(*externalhttp.Response) (int, string)

var minRequestDuration time.Duration

// Init is a workflow function that initialises the workflow based on the given flags
func Init(flags Flags) {
	millisecondsPerRequest := 1000 / flags.Rate
	minRequestDuration = time.Duration(millisecondsPerRequest) * time.Millisecond
}

// Run is a workflow function that orchestrates the API testing
func Run(definition definition.Definition) {

	apiSummary := apiSummary{
		baseURI:       definition.BaseURI,
		authDetails:   definition.AuthDetails,
		globalHeaders: definition.API.GlobalHeaders,
		globalMethods: definition.API.GlobalMethods,
	}

	for endpoint, endpointDetails := range definition.API.Endpoints {
		requestOptions := buildEndpointRequestOptions(apiSummary, endpoint, endpointDetails)
		requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)

		testEndpointWithAllLevelsOfAuthentication(requestOptions, apiSummary.authDetails, "token")
	}
}

func testEndpointWithAllLevelsOfAuthentication(requestOptions http.RequestOptions,
	authDetails definition.AuthDetails, testNamePrefix string) {

	testEndpointWithAuthToken(requestOptions, authDetails, testNamePrefix+"-high",
		authDetails.High, verifyResponseExpectedOK)
	testEndpointWithAuthToken(requestOptions, authDetails, testNamePrefix+"-low",
		authDetails.Low, verifyResponseExpectedUnauthorised)
	testEndpointWithoutAuthToken(requestOptions, authDetails, testNamePrefix+"-none",
		verifyResponseExpectedUnauthorised)
}

func testEndpointWithAuthToken(requestOptions http.RequestOptions,
	authDetails definition.AuthDetails, testName string, token string, verifier verifier) {

	logger.TestPrefix(requestOptions.Endpoint, testName)
	authHeaderValue := buildAuthHeaderValue(authDetails.HeaderValuePrefix, token)
	requestOptions.Headers = addHeader(requestOptions.Headers, authDetails.HeaderName, authHeaderValue)

	testEndpoint(requestOptions, verifier)
}

func testEndpointWithoutAuthToken(requestOptions http.RequestOptions,
	authDetails definition.AuthDetails, testName string, verifier verifier) {

	logger.TestPrefix(requestOptions.Endpoint, testName)
	requestOptions.Headers = removeHeader(requestOptions.Headers, authDetails.HeaderName)

	testEndpoint(requestOptions, verifier)
}

func testEndpoint(requestOptions http.RequestOptions, verifier verifier) {
	startTime := time.Now()
	response, err := buildAndSendRequest(requestOptions)
	if err != nil {
		logger.Message("Skipping due to error: " + err.Error())
		return
	}

	status, result := verifier(response)
	logger.TestResult(strconv.Itoa(status) + " " + result)

	if result != "OK" {
		logger.DumpRequest(requestOptions)
	}

	durationSinceStartTime := time.Since(startTime)
	time.Sleep(minRequestDuration - durationSinceStartTime)
}
