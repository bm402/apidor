package main

import (
	"flag"
	"fmt"

	"github.com/bncrypted/apidor/pkg/definition"
)

func main() {
	definitionFile := flag.String("d", "definitions/sample.yml", "Path to the API definition YAML file")
	flag.Parse()
	definitions := definition.ReadDefinition(*definitionFile)
	fmt.Println(definitions)
}
