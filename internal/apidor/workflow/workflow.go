package workflow

import (
	externalhttp "net/http"
	"strconv"
	"time"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/internal/apidor/testcode"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

// Flags is a workflow struct that holds command line flags for customising the workflow
type Flags struct {
	Rate      int
	TestCodes testcode.TestCodes
}

type apiSummary struct {
	baseURI       string
	authDetails   definition.AuthDetails
	globalHeaders map[string]string
	globalMethods []string
}

type verifier func(*externalhttp.Response, []string) (int, string)

var requestID int

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

	testCodes := flags.TestCodes
	isRunAllTests := testCodes.Contains(testcode.ALL)

	requestID = 1

	// wrapped vars for parameter wrapping test
	wrappedVarsInArrays, wrappedVarsInMaps := wrapVars(definition.Vars)

	for endpoint, endpointOperations := range definition.API.Endpoints {
		unusedEndpointMethods := findUnusedEndpointMethods(definition.API.GlobalMethods, endpointOperations)

		for _, endpointOperationDetails := range endpointOperations {
			baseRequestOptions := buildEndpointRequestOptions(apiSummary, endpoint, endpointOperationDetails)
			bannedResponseWords := getHighPrivilegedVariableValues(baseRequestOptions, definition.Vars)
			var requestOptions http.RequestOptions
			var collectedRequestOptions []http.RequestOptions

			// high privileged request
			if isRunAllTests || testCodes.Contains(testcode.HP) {
				if endpointOperationDetails.IsDeleteOperation {
					logger.TestPrefix(requestID, endpoint, "high-priv")
					logger.TestResult("Skipping delete operation")
				} else {
					requestOptions = baseRequestOptions.DeepCopy()
					requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
						definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.High)
					requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
					logger.TestPrefix(requestID, endpoint, "high-priv")
					testEndpoint(requestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)
				}
			}

			// low privileged requests
			if isRunAllTests || testCodes.Contains(testcode.LP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
				collectedRequestOptions = substituteMixedPrivilegedVariablePermutations(requestOptions, definition.Vars)
				for _, requestOptions := range collectedRequestOptions {
					logger.TestPrefix(requestID, endpoint, "low-priv-perms")
					testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
						bannedResponseWords, minRequestDuration)
				}
			}

			// parameter pollution in request params
			if isRunAllTests || testCodes.Contains(testcode.RPP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
				collectedRequestOptions = substituteAndParameterPolluteRequestParams(requestOptions, definition.Vars)
				for _, requestOptions := range collectedRequestOptions {
					logger.TestPrefix(requestID, endpoint, "low-priv-request-pp")
					testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
						bannedResponseWords, minRequestDuration)
				}
			}

			// parameter pollution in body params
			if isRunAllTests || testCodes.Contains(testcode.BPP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
				collectedRequestOptions = substituteAndParameterPolluteBodyParams(requestOptions, definition.Vars)
				for _, requestOptions := range collectedRequestOptions {
					logger.TestPrefix(requestID, endpoint, "low-priv-body-pp")
					testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
						bannedResponseWords, minRequestDuration)
				}
			}

			// parameter wrapping in body params
			if isRunAllTests || testCodes.Contains(testcode.PW) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
				collectedRequestOptions = substituteAndParameterWrapBodyParams(requestOptions, definition.Vars,
					wrappedVarsInArrays, wrappedVarsInMaps)
				for _, requestOptions := range collectedRequestOptions {
					logger.TestPrefix(requestID, endpoint, "low-priv-body-pw")
					testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
						bannedResponseWords, minRequestDuration)
				}
			}

			// method replacement
			if isRunAllTests || testCodes.Contains(testcode.MR) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
				requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
				collectedRequestOptions = substituteUnusedMethods(requestOptions, unusedEndpointMethods)
				for _, requestOptions := range collectedRequestOptions {
					logger.TestPrefix(requestID, endpoint, "low-priv-msub")
					testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
						bannedResponseWords, minRequestDuration)
				}
			}

			// no privilege request
			if isRunAllTests || testCodes.Contains(testcode.NP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = removeAuthHeaderFromRequestOptions(requestOptions, definition.AuthDetails.HeaderName)
				requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
				logger.TestPrefix(requestID, endpoint, "no-priv")
				testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
					bannedResponseWords, minRequestDuration)
			}
		}
	}
}

func testEndpoint(requestOptions http.RequestOptions, verifier verifier,
	bannedResponseWords []string, minRequestDuration time.Duration) {

	startTime := time.Now()
	requestOptions.Headers["X-Apidor-Request-ID"] = strconv.Itoa(requestID)
	requestID++

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
