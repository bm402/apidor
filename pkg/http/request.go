package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

// RequestOptions holds the data that builds the HTTP request
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

// Request is a method for sending a HTTP request
func Request(requestOptions RequestOptions) *http.Response {
	uri := buildURI(requestOptions.BaseURI, requestOptions.Endpoint, requestOptions.RequestParams)
	body := buildBody(requestOptions.ContentType, requestOptions.BodyParams)
	contentType := buildContentType(requestOptions.ContentType)

	timeout := time.Duration(5 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest(requestOptions.Method, uri, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	for headerName, headerValue := range requestOptions.Headers {
		req.Header.Set(headerName, headerValue)
	}
	req.Header.Set("Content-Type", contentType)

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(reqDump))

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}
