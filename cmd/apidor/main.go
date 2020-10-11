package main

import (
	"flag"
	"os"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/internal/apidor/workflow"
	"github.com/bncrypted/apidor/pkg/definition"
	"github.com/bncrypted/apidor/pkg/http"
)

func main() {
	definitionFile := flag.String("d", "definitions/sample.yml", "Path to the API definition YAML file")
	localCertFile := flag.String("cert", "", "Path to a local certificate authority file")
	logFile := flag.String("o", "", "Log file name")
	proxyURI := flag.String("proxy", "", "Gives a URI to proxy HTTP traffic through")
	rate := flag.Int("rate", 5, "Specifies maximum number of requests made per second")
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

	workflowFlags := workflow.Flags{
		Rate: *rate,
	}
	workflow.Run(definition, workflowFlags)

	logger.Finished()
}
