package definition

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Definition is a full model of the YAML definition file
type Definition struct {
	BaseURI     string           `yaml:"base"`
	AuthDetails AuthDetails      `yaml:"auth"`
	Vars        map[string][]int `yaml:"vars"`
	API         APIDetails       `yaml:"api"`
}

// AuthDetails is a model of the authentication header details
type AuthDetails struct {
	HeaderName        string `yaml:"header_name"`
	HeaderValuePrefix string `yaml:"header_value_prefix"`
	High              string `yaml:"high_privileged_access_token"`
	Low               string `yaml:"low_privileged_access_token"`
}

// APIDetails is a model for the API details
type APIDetails struct {
	GlobalMethods []string                   `yaml:"methods"`
	GlobalHeaders map[string]string          `yaml:"headers"`
	Endpoints     map[string]EndpointDetails `yaml:"endpoints"`
}

// EndpointDetails is a model for an endpoint in the API
type EndpointDetails struct {
	Operation     string                 `yaml:"operation"`
	Method        string                 `yaml:"method"`
	ContentType   string                 `yaml:"content_type"`
	Headers       map[string]string      `yaml:"headers"`
	RequestParams map[string]string      `yaml:"request_params"`
	BodyParams    map[string]interface{} `yaml:"body_params"`
}

// Read is a function that reads the API definition from the given YAML file
func Read(filepath string) Definition {
	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	var definition Definition
	err = yaml.Unmarshal(buf, &definition)
	if err != nil {
		panic(err)
	}

	return definition
}
