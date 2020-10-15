package workflow

import (
	externalhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/internal/apidor/testcode"
	model "github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

// Flags is a workflow struct that holds command line flags for customising the workflow
type Flags struct {
	EndpointToTest string
	Rate           int
	TestCodes      testcode.TestCodes
}

type apiSummary struct {
	baseURI       string
	authDetails   model.AuthDetails
	globalHeaders map[string]string
	globalMethods []string
}

type verifier func(*externalhttp.Response, []string) (int, string)

var requestID int

// Run is a workflow function that orchestrates the API testing
func Run(definition model.Definition, flags Flags) {

	apiSummary := apiSummary{
		baseURI:       definition.BaseURI,
		authDetails:   definition.AuthDetails,
		globalHeaders: definition.API.GlobalHeaders,
		globalMethods: definition.API.GlobalMethods,
	}

	millisecondsPerRequest := 1000 / flags.Rate
	minRequestDuration := time.Duration(millisecondsPerRequest) * time.Millisecond

	if flags.EndpointToTest != "all" {
		if endpointName, endpointOperationDetails,
			ok := verifyEndpoint(flags.EndpointToTest, definition.API.Endpoints); ok {
			definition.API.Endpoints = map[string][]model.EndpointDetails{
				endpointName: []model.EndpointDetails{endpointOperationDetails}}
		} else {
			logger.Fatal("Could not find endpoint to test \"" + flags.EndpointToTest + "\"")
			return
		}
	}

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

			// request parameter substitution
			if isRunAllTests || testCodes.Contains(testcode.RPS) || testCodes.Contains(testcode.RPSPP) {
				if isRunAllTests || testCodes.Contains(testcode.RPSPP) {
					requestOptions = baseRequestOptions.DeepCopy()
					requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
						definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
					collectedRequestOptions = substituteAndMoveAndParameterPolluteBodyParamsToRequestParams(
						requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpoint, "low-priv-rps-pp")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					requestOptions = baseRequestOptions.DeepCopy()
					requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
						definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
					collectedRequestOptions = substituteAndMoveBodyParamsToRequestParams(requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpoint, "low-priv-rps")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
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

			// append .json
			if isRunAllTests || testCodes.Contains(testcode.JSON) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)
				requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
				requestOptions.Endpoint += ".json"
				logger.TestPrefix(requestID, endpoint, "low-priv-json")
				testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
					bannedResponseWords, minRequestDuration)
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

func verifyEndpoint(endpointCode string, endpoints map[string][]model.EndpointDetails) (string,
	model.EndpointDetails, bool) {

	endpointCodeComponents := strings.Split(endpointCode, " ")
	if len(endpointCodeComponents) != 2 {
		return "", model.EndpointDetails{}, false
	}

	if endpointDetails, ok := endpoints[endpointCodeComponents[1]]; ok {
		for _, endpointOperationDetails := range endpointDetails {
			if endpointOperationDetails.Method == endpointCodeComponents[0] {
				return endpointCodeComponents[1], endpointOperationDetails, true
			}
		}
	}

	return "", model.EndpointDetails{}, false
}
