package main

import (
	"flag"
	"fmt"
	"net/http/httputil"

	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func main() {
	definitionFile := flag.String("d", "definitions/sample.yml", "Path to the API definition YAML file")
	flag.Parse()

	def := definition.Read(*definitionFile)

	endpoint := def.API.Endpoints["users.info"]
	headers := make(map[string]string)
	for headerName, headerValue := range endpoint.Headers {
		headers[headerName] = headerValue
	}
	headers["Authorization"] = def.AuthDetails.HeaderValuePrefix + " " + def.AuthDetails.High

	requestOptions := http.RequestOptions{
		Method:        endpoint.Method,
		BaseURI:       def.BaseURI,
		Endpoint:      "users.info",
		ContentType:   endpoint.ContentType,
		Headers:       headers,
		RequestParams: endpoint.RequestParams,
		BodyParams:    endpoint.BodyParams,
	}

	resp := http.Request(requestOptions)
	defer resp.Body.Close()

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(respDump))
}
