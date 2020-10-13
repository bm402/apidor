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

type verifier func(*externalhttp.Response, []string) (int, string)

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
		baseRequestOptions := buildEndpointRequestOptions(apiSummary, endpoint, endpointDetails)
		bannedResponseWords := getHighPrivilegedVariableValues(baseRequestOptions, definition.Vars)

		// high privileged request
		// requestOptions := baseRequestOptions.DeepCopy()
		// requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
		// 	definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.High)
		// requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
		// logger.TestPrefix(endpoint, "high-priv")
		// testEndpoint(requestOptions, verifyResponseExpectedOK, []string{})

		// // low privileged requests
		// requestOptions = baseRequestOptions.DeepCopy()
		// requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
		// 	definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
		// collectedRequestOptions := substituteAllPrivilegedVariablePermutations(requestOptions, definition.Vars)
		// for _, requestOptions := range collectedRequestOptions {
		// 	logger.TestPrefix(endpoint, "low-priv-permutations")
		// 	testEndpoint(requestOptions, verifyResponseExpectedUnauthorised, bannedResponseWords)
		// }

		// // parameter pollution in request params
		// requestOptions := baseRequestOptions.DeepCopy()
		// requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
		// 	definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
		// collectedRequestOptions := substituteOppositePrivilegedRequestParamPermutations(requestOptions, definition.Vars)
		// for _, requestOptions := range collectedRequestOptions {
		// 	logger.TestPrefix(endpoint, "low-priv-request-pp")
		// 	testEndpoint(requestOptions, verifyResponseExpectedUnauthorised, bannedResponseWords)
		// }

		// parameter pollution in body params
		requestOptions := baseRequestOptions.DeepCopy()
		requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
			definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
		collectedRequestOptions := substituteOppositePrivilegedBodyParamPermutations(requestOptions, definition.Vars)
		for _, requestOptions := range collectedRequestOptions {
			logger.TestPrefix(endpoint, "low-priv-body-pp")
			testEndpoint(requestOptions, verifyResponseExpectedUnauthorised, bannedResponseWords)
		}

		// // no privilege request
		// requestOptions = baseRequestOptions.DeepCopy()
		// requestOptions = removeAuthHeaderFromRequestOptions(requestOptions, definition.AuthDetails.HeaderName)
		// requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
		// logger.TestPrefix(endpoint, "no-priv")
		// testEndpoint(requestOptions, verifyResponseExpectedUnauthorised, bannedResponseWords)
	}

}

func testEndpoint(requestOptions http.RequestOptions, verifier verifier, bannedResponseWords []string) {
	startTime := time.Now()
	response, err := buildAndSendRequest(requestOptions)
	if err != nil {
		logger.Message("Skipping due to error: " + err.Error())
		return
	}

	status, result := verifier(response, bannedResponseWords)
	logger.TestResult(strconv.Itoa(status) + " " + result)

	if result[:2] != "OK" {
		logger.DumpRequest(requestOptions)
	}

	durationSinceStartTime := time.Since(startTime)
	time.Sleep(minRequestDuration - durationSinceStartTime)
}
