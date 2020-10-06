package definition

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Definition is a full model of the YAML definition file
type Definition struct {
	baseURI     string                   `yaml:"base"`
	authDetails AuthDetails              `yaml:"auth"`
	vars        map[string][]interface{} `yaml:"vars"`
	api         APIDetails               `yaml:"api"`
}

// AuthDetails is a model of the authentication header details
type AuthDetails struct {
	headerName        string `yaml:"header_name"`
	headerValuePrefix string `yaml:"header_value_prefix"`
	high              string `yaml:"high_privileged_access_token"`
	low               string `yaml:"low_privileged_access_token"`
}

// APIDetails is a model for the API details
type APIDetails struct {
	globalMethods []string                   `yaml:"methods"`
	globalHeaders map[string]string          `yaml:"headers"`
	endpoints     map[string]EndpointDetails `yaml:"endpoints"`
}

// EndpointDetails is a model for an endpoint in the API
type EndpointDetails struct {
	operation     string                 `yaml:"operation"`
	method        string                 `yaml:"method"`
	contentType   string                 `yaml:"content_type"`
	headers       map[string]string      `yaml:"headers"`
	requestParams map[string]string      `yaml:"request_params"`
	bodyParams    map[string]interface{} `yaml:"body_params"`
}

// ReadDefinition reads the API definition from the given YAML file
func ReadDefinition(path string) Definition {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading definition file")
	}

	var definition Definition
	err = yaml.Unmarshal(buf, &definition)
	if err != nil {
		fmt.Println("YAML parsing error")
	}

	return definition
}
