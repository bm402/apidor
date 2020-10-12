package workflow

import (
	externalhttp "net/http"
	"net/http/httputil"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func buildEndpointRequestOptions(apiSummary apiSummary, endpointName string,
	endpointDetails definition.EndpointDetails) http.RequestOptions {
	headers := mergeGlobalAndLocalHeaders(apiSummary.globalHeaders, endpointDetails.Headers)
	requestOptions := http.RequestOptions{
		Method:        endpointDetails.Method,
		BaseURI:       apiSummary.baseURI,
		Endpoint:      endpointName,
		ContentType:   endpointDetails.ContentType,
		Headers:       headers,
		RequestParams: endpointDetails.RequestParams,
		BodyParams:    endpointDetails.BodyParams,
	}
	return requestOptions
}

func buildAndSendRequest(requestOptions http.RequestOptions) (*externalhttp.Response, error) {
	request, err := http.CreateRequest(requestOptions)
	if err != nil {
		return nil, err
	}

	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		logger.DebugError(err.Error())
	} else {
		logger.DebugMessage(string(requestDump))
	}

	response, err := http.SendRequest(request)
	if err != nil {
		return nil, err
	}

	responseDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		logger.DebugError(err.Error())
	} else {
		logger.DebugMessage(string(responseDump))
	}

	return response, nil
}

func mergeGlobalAndLocalHeaders(globalHeaders map[string]string, localHeaders map[string]string) map[string]string {
	headers := make(map[string]string)
	for headerName, headerValue := range globalHeaders {
		headers[headerName] = headerValue
	}
	for headerName, headerValue := range localHeaders {
		headers[headerName] = headerValue
	}
	return headers
}

func addAuthHeaderToRequestOptions(requestOptions http.RequestOptions, headerName string,
	headerValuePrefix string, token string) http.RequestOptions {

	headerValue := buildAuthHeaderValue(headerValuePrefix, token)
	requestOptions.Headers = addHeader(requestOptions.Headers, headerName, headerValue)
	return requestOptions
}

func removeAuthHeaderFromRequestOptions(requestOptions http.RequestOptions, headerName string) http.RequestOptions {
	requestOptions.Headers = removeHeader(requestOptions.Headers, headerName)
	return requestOptions
}

func addHeader(headers map[string]string, headerName string, headerValue string) map[string]string {
	headers[headerName] = headerValue
	return headers
}

func removeHeader(headers map[string]string, headerName string) map[string]string {
	delete(headers, headerName)
	return headers
}

func buildAuthHeaderValue(prefix string, token string) string {
	if prefix == "" {
		return token
	}
	return prefix + " " + token
}
