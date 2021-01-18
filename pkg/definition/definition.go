package definition

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Definition is a full model of the YAML definition file
type Definition struct {
	BaseURI     string               `yaml:"base"`
	AuthDetails AuthDetails          `yaml:"auth"`
	Vars        map[string]Variables `yaml:"vars"`
	API         APIDetails           `yaml:"api"`
}

// AuthDetails is a model of the authentication header details
type AuthDetails struct {
	HeaderName        string `yaml:"header_name"`
	HeaderValuePrefix string `yaml:"header_value_prefix"`
	High              string `yaml:"high_privileged_access_token"`
	Low               string `yaml:"low_privileged_access_token"`
}

// Variables is a model of high and low privileged variables
type Variables struct {
	High  interface{} `yaml:"high"`
	Low   interface{} `yaml:"low"`
	Alias string      `yaml:"alias"`
}

// APIDetails is a model for the API details
type APIDetails struct {
	GlobalMethods []string                     `yaml:"methods"`
	GlobalHeaders map[string]string            `yaml:"headers"`
	Endpoints     map[string][]EndpointDetails `yaml:"endpoints"`
}

// EndpointDetails is a model for an endpoint in the API
type EndpointDetails struct {
	Method            string                 `yaml:"method"`
	IsDeleteOperation bool                   `yaml:"is_delete"`
	ContentType       string                 `yaml:"content_type"`
	Headers           map[string]string      `yaml:"headers"`
	RequestParams     map[string]string      `yaml:"request_params"`
	BodyParams        map[string]interface{} `yaml:"body_params"`
}

// Read is a function that reads the API definition from the given YAML file
func Read(filepath string) (Definition, error) {
	var definition Definition

	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		return definition, err
	}

	err = yaml.Unmarshal(buf, &definition)
	if err != nil {
		return definition, err
	}

	definition.setDefaultValues()
	definition.normaliseBodyParamMaps()
	return definition, nil
}

func (d *Definition) setDefaultValues() {
	for varName, varValues := range d.Vars {
		if varValues.Alias == "" {
			varValues.Alias = varName
			d.Vars[varName] = varValues
		}
	}
}

func (d *Definition) normaliseBodyParamMaps() {
	for endpointName, endpointDetails := range d.API.Endpoints {
		for endpointOperationIdx, endpointOperationDetails := range endpointDetails {
			bodyParams := make(map[string]interface{})
			for paramName, paramValue := range endpointOperationDetails.BodyParams {
				switch paramValue.(type) {
				case []interface{}:
					bodyParams[paramName] = normaliseArray(paramValue.([]interface{}))
				case map[interface{}]interface{}:
					bodyParams[paramName] = normaliseMap(paramValue.(map[interface{}]interface{}))
				default:
					bodyParams[paramName] = paramValue
				}
			}
			d.API.Endpoints[endpointName][endpointOperationIdx].BodyParams = bodyParams
		}
	}
}

func normaliseMap(mp map[interface{}]interface{}) map[string]interface{} {
	normalisedMap := make(map[string]interface{})
	for key, value := range mp {
		switch value.(type) {
		case []interface{}:
			normalisedMap[fmt.Sprintf("%v", key)] = normaliseArray(value.([]interface{}))
		case map[interface{}]interface{}:
			normalisedMap[fmt.Sprintf("%v", key)] = normaliseMap(value.(map[interface{}]interface{}))
		default:
			normalisedMap[fmt.Sprintf("%v", key)] = value
		}
	}
	return normalisedMap
}

func normaliseArray(arr []interface{}) []interface{} {
	normalisedArr := make([]interface{}, len(arr))
	for idx, value := range arr {
		switch value.(type) {
		case []interface{}:
			normalisedArr[idx] = normaliseArray(value.([]interface{}))
		case map[interface{}]interface{}:
			normalisedArr[idx] = normaliseMap(value.(map[interface{}]interface{}))
		default:
			normalisedArr[idx] = value
		}
	}
	return normalisedArr
}
