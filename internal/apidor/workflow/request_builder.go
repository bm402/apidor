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

func buildAndSendRequest(requestOptions http.RequestOptions) *externalhttp.Response {
	request := http.CreateRequest(requestOptions)
	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		panic(err)
	}
	logger.DebugMessage(string(requestDump))

	response := http.SendRequest(request)

	responseDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		panic(err)
	}
	logger.DebugMessage(string(responseDump))

	return response
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
