package main

import (
	"flag"
	"net/http/httputil"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func main() {
	flags := logger.Flags{
		DefinitionFile: flag.String("d", "definitions/sample.yml", "Path to the API definition YAML file"),
		LogFile:        flag.String("o", "", "Log file name"),
		IsDebug:        flag.Bool("debug", false, "Specifies whether to use debugging mode for verbose output"),
	}
	flag.Parse()

	logger.Init(flags)
	defer logger.Close()
	logger.Logo()

	def := definition.Read(*flags.DefinitionFile)

	logger.RunInfo(def.BaseURI, len(def.API.Endpoints), flags)
	logger.Starting()

	logger.TestPrefix("users.info", "happy-path")
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

	req := http.CreateRequest(requestOptions)
	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		panic(err)
	}
	logger.DebugMessage(string(reqDump))

	resp := http.SendRequest(req)
	defer resp.Body.Close()

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}
	logger.DebugMessage(string(respDump))

	logger.TestResult(resp.Status)

	logger.Finished()
}
