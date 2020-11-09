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
	EndpointToTest   string
	Rate             int
	TestCodes        testcode.TestCodes
	IsIgnoreBaseCase bool
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
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "high-priv")
					logger.TestResult("Skipping delete operation")
				} else {
					requestOptions = baseRequestOptions.DeepCopy()
					requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
						definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.High)
					requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "high-priv")
					testEndpoint(requestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)
				}
			}

			// low privileged requests
			if isRunAllTests || testCodes.Contains(testcode.LP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				baseCaseRequestOptions := substituteLowPrivilegedVariables(requestOptions, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-perms-base")
				baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedVariables(requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-perms")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-perms]")
				}
			}

			// parameter pollution in request params
			if isRunAllTests || testCodes.Contains(testcode.RPP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				// normal parameter pollution
				baseCaseRequestOptions := substituteLowPrivilegedAndParameterPolluteRequestParams(
					requestOptions, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pp-base")
				baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndParameterPolluteRequestParams(
						requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pp")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-request-pp]")
				}

				// square bracket parameter pollution
				requestOptionsWithSquareBrackets := requestOptions.DeepCopy()
				requestParamsWithSquareBrackets := make(map[string]string)
				for paramName, paramValue := range requestOptionsWithSquareBrackets.RequestParams {
					requestParamsWithSquareBrackets[paramName+"[]"] = paramValue
				}
				requestOptionsWithSquareBrackets.RequestParams = requestParamsWithSquareBrackets

				baseCaseRequestOptions = substituteLowPrivilegedAndParameterPolluteRequestParams(
					requestOptionsWithSquareBrackets, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pp-sb-base")
				baseStatus = testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndParameterPolluteRequestParams(
						requestOptionsWithSquareBrackets, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pp-sb")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-request-pp-sb]")
				}
			}

			// parameter pollution in body params
			if isRunAllTests || testCodes.Contains(testcode.BPP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				baseCaseRequestOptions := substituteLowPrivilegedAndParameterPolluteBodyParams(
					requestOptions, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-body-pp-base")
				baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndParameterPolluteBodyParams(
						requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-body-pp")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-body-pp]")
				}
			}

			// parameter wrapping in request params
			if isRunAllTests || testCodes.Contains(testcode.RPW) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				// wrap array in square brackets
				requestOptionsWithSquareBrackets := requestOptions.DeepCopy()
				requestParamsWithSquareBrackets := make(map[string]string)
				for paramName, paramValue := range requestOptionsWithSquareBrackets.RequestParams {
					requestParamsWithSquareBrackets[paramName+"[]"] = paramValue
				}
				requestOptionsWithSquareBrackets.RequestParams = requestParamsWithSquareBrackets

				baseCaseRequestOptions := substituteLowPrivilegedVariables(
					requestOptionsWithSquareBrackets, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pw-arr-sb-base")
				baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedVariables(
						requestOptionsWithSquareBrackets, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pw-arr-sb")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-request-pw-arr-sb]")
				}

				// wrap in object in square brackets
				requestOptionsWithSquareBrackets = requestOptions.DeepCopy()
				requestParamsWithSquareBrackets = make(map[string]string)
				for paramName, paramValue := range requestOptionsWithSquareBrackets.RequestParams {
					requestParamsWithSquareBrackets[paramName+"["+paramName+"]"] = paramValue
				}
				requestOptionsWithSquareBrackets.RequestParams = requestParamsWithSquareBrackets

				baseCaseRequestOptions = substituteLowPrivilegedVariables(
					requestOptionsWithSquareBrackets, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pw-obj-sb-base")
				baseStatus = testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedVariables(
						requestOptionsWithSquareBrackets, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pw-obj-sb")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-request-pw-obj-sb]")
				}

				// wrap in object in dot notation
				requestOptionsWithDotNotation := requestOptions.DeepCopy()
				requestParamsWithDotNotation := make(map[string]string)
				for paramName, paramValue := range requestOptionsWithDotNotation.RequestParams {
					requestParamsWithDotNotation[paramName+"."+paramName] = paramValue
				}
				requestOptionsWithDotNotation.RequestParams = requestParamsWithDotNotation

				baseCaseRequestOptions = substituteLowPrivilegedVariables(
					requestOptionsWithDotNotation, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pw-obj-dn-base")
				baseStatus = testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK, []string{}, minRequestDuration)

				if is2xx(baseStatus) || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedVariables(
						requestOptionsWithDotNotation, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-request-pw-obj-dn")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase case failed, skipping rest of [low-priv-request-pw-obj-dn]")
				}
			}

			// parameter wrapping in body params
			if isRunAllTests || testCodes.Contains(testcode.BPW) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				// wrap in arrays
				collectedBaseCaseRequestOptions := substituteLowPrivilegedAndParameterWrapBodyParams(
					requestOptions, definition.Vars, wrappedVarsInArrays)
				isBaseStatus2xx := false

				for _, baseCaseRequestOptions := range collectedBaseCaseRequestOptions {
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-body-pw-arr-base")
					baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK,
						[]string{}, minRequestDuration)
					if is2xx(baseStatus) {
						isBaseStatus2xx = true
						break
					}
				}

				if isBaseStatus2xx || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndParameterWrapBodyParams(
						requestOptions, definition.Vars, wrappedVarsInArrays)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-body-pw-arr")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase cases failed, skipping rest of [low-priv-body-pw-arr]")
				}

				// wrap in objects
				collectedBaseCaseRequestOptions = substituteLowPrivilegedAndParameterWrapBodyParams(
					requestOptions, definition.Vars, wrappedVarsInMaps)
				isBaseStatus2xx = false

				for _, baseCaseRequestOptions := range collectedBaseCaseRequestOptions {
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-body-pw-obj-base")
					baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK,
						[]string{}, minRequestDuration)
					if is2xx(baseStatus) {
						isBaseStatus2xx = true
						break
					}
				}

				if isBaseStatus2xx || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndParameterWrapBodyParams(
						requestOptions, definition.Vars, wrappedVarsInMaps)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-body-pw-obj")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase cases failed, skipping rest of [low-priv-body-pw-obj]")
				}
			}

			// request parameter substitution
			if isRunAllTests || testCodes.Contains(testcode.RPSPP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				collectedBaseCaseRequestOptions := substituteLowPrivilegedAndParameterPolluteBodyParamsToRequestParams(
					requestOptions, definition.Vars)
				isBaseStatus2xx := false

				for _, baseCaseRequestOptions := range collectedBaseCaseRequestOptions {
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-rps-pp-base")
					baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK,
						[]string{}, minRequestDuration)
					if is2xx(baseStatus) {
						isBaseStatus2xx = true
						break
					}
				}

				if isBaseStatus2xx || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndParameterPolluteBodyParamsToRequestParams(
						requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-rps-pp")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase cases failed, skipping rest of [low-priv-rps-pp]")
				}

			} else if testCodes.Contains(testcode.RPS) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = addAuthHeaderToRequestOptions(requestOptions, definition.AuthDetails.HeaderName,
					definition.AuthDetails.HeaderValuePrefix, definition.AuthDetails.Low)

				collectedBaseCaseRequestOptions := substituteLowPrivilegedAndMoveBodyParamsToRequestParams(
					requestOptions, definition.Vars)
				isBaseStatus2xx := false

				for _, baseCaseRequestOptions := range collectedBaseCaseRequestOptions {
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-rps-base")
					baseStatus := testEndpoint(baseCaseRequestOptions, verifyResponseExpectedOK,
						[]string{}, minRequestDuration)
					if is2xx(baseStatus) {
						isBaseStatus2xx = true
						break
					}
				}

				if isBaseStatus2xx || flags.IsIgnoreBaseCase {
					collectedRequestOptions = substituteMixedPrivilegedAndMoveBodyParamsToRequestParams(
						requestOptions, definition.Vars)
					for _, requestOptions := range collectedRequestOptions {
						logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-rps")
						testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
							bannedResponseWords, minRequestDuration)
					}
				} else {
					logger.Message("\tBase cases failed, skipping rest of [low-priv-rps]")
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
					logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-msub")
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
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "low-priv-json")
				testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
					bannedResponseWords, minRequestDuration)
			}

			// no privilege request
			if isRunAllTests || testCodes.Contains(testcode.NP) {
				requestOptions = baseRequestOptions.DeepCopy()
				requestOptions = removeAuthHeaderFromRequestOptions(requestOptions, definition.AuthDetails.HeaderName)
				requestOptions = substituteHighPrivilegedVariables(requestOptions, definition.Vars)
				logger.TestPrefix(requestID, endpointOperationDetails.Method, endpoint, "no-priv")
				testEndpoint(requestOptions, verifyResponseExpectedUnauthorised,
					bannedResponseWords, minRequestDuration)
			}
		}
	}
}

func testEndpoint(requestOptions http.RequestOptions, verifier verifier,
	bannedResponseWords []string, minRequestDuration time.Duration) int {

	startTime := time.Now()
	requestOptions.Headers["X-Apidor-Request-ID"] = strconv.Itoa(requestID)
	requestID++

	response, err := buildAndSendRequest(requestOptions)
	if err != nil {
		logger.Message("Skipping due to error: " + err.Error())
		return 0
	}

	status, result := verifier(response, bannedResponseWords)
	logger.TestResult(strconv.Itoa(status) + " " + result)

	// if result[:2] != "OK" {
	// 	logger.DumpRequest(requestOptions)
	// }

	durationSinceStartTime := time.Since(startTime)
	time.Sleep(minRequestDuration - durationSinceStartTime)

	return status
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

func is2xx(status int) bool {
	return status/100 == 2
}
