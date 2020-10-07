package http

import (
	"bytes"
	"net/http"
	"time"
)

// RequestOptions is a http struct that holds the data that builds the HTTP request
type RequestOptions struct {
	Method        string
	BaseURI       string
	Endpoint      string
	ContentType   string
	Headers       map[string]string
	RequestParams map[string]string
	BodyParams    map[string]interface{}
	IsProxy       bool
	ProxyURI      string
}

// CreateRequest is a http function for creating a HTTP request based on request options
func CreateRequest(requestOptions RequestOptions) *http.Request {
	uri := buildURI(requestOptions.BaseURI, requestOptions.Endpoint, requestOptions.RequestParams)
	body := buildBody(requestOptions.ContentType, requestOptions.BodyParams)
	contentType := buildContentType(requestOptions.ContentType)

	req, err := http.NewRequest(requestOptions.Method, uri, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	for headerName, headerValue := range requestOptions.Headers {
		req.Header.Set(headerName, headerValue)
	}
	req.Header.Set("Content-Type", contentType)

	return req
}

// SendRequest is a http function for sending a HTTP request and returns the response
func SendRequest(req *http.Request) *http.Response {
	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}
