package main

import (
	"flag"
	"os"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/internal/apidor/testcode"
	"github.com/bncrypted/apidor/internal/apidor/workflow"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func main() {
	localCertFile := flag.String("cert", "", "Path to a local certificate authority file")
	definitionFile := flag.String("d", "definitions/sample.yml", "Path to the API definition YAML file")
	endpoint := flag.String("e", "all", "Specifies a single endpoint operation to test")
	logFile := flag.String("o", "", "Log file name")
	proxyURI := flag.String("proxy", "", "Gives a URI to proxy HTTP traffic through")
	rate := flag.Int("rate", 5, "Specifies maximum number of requests made per second")
	tests := flag.String("tests", "all", "Specifies which tests should be executed")
	isDebug := flag.Bool("debug", false, "Specifies whether to use debugging mode for verbose output")
	flag.Parse()

	httpFlags := http.Flags{
		ProxyURI:      *proxyURI,
		LocalCertFile: *localCertFile,
	}

	http.Init(httpFlags)

	loggerFlags := logger.Flags{
		DefinitionFile: *definitionFile,
		LocalCertFile:  *localCertFile,
		LogFile:        *logFile,
		ProxyURI:       *proxyURI,
		Rate:           *rate,
		IsDebug:        *isDebug,
	}

	logger.Init(loggerFlags)
	defer logger.Close()
	logger.Logo()

	definition, err := definition.Read(*definitionFile)
	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}

	logger.RunInfo(definition.BaseURI, len(definition.API.Endpoints), loggerFlags)
	logger.Starting()

	testCodes, err := testcode.ParseTestCodes(*tests)
	if err != nil {
		logger.Error(err.Error())
	}

	workflowFlags := workflow.Flags{
		EndpointToTest: *endpoint,
		Rate:           *rate,
		TestCodes:      testCodes,
	}
	workflow.Run(definition, workflowFlags)

	logger.Finished()
}
