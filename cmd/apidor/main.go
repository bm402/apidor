package main

import (
	"flag"
	"os"

	"github.com/bncrypted/apidor/internal/apidor/logger"
	"github.com/bncrypted/apidor/internal/apidor/workflow"
	"github.com/bncrypted/apidor/pkg/definition"
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

	definition, err := definition.Read(*flags.DefinitionFile)
	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}

	logger.RunInfo(definition.BaseURI, len(definition.API.Endpoints), flags)
	logger.Starting()

	workflow.Run(definition, workflow.Flags{})

	logger.Finished()
}
