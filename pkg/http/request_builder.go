package http

import (
	"encoding/json"
)

func buildURI(baseURI string, endpoint string, requestParams map[string]string) string {
	uri := ""

	if baseURI[len(baseURI)-1] == '/' && endpoint[0] == '/' {
		// removes potential double slash between baseURI and endpoint
		uri += baseURI + endpoint[1:]
	} else if baseURI[len(baseURI)-1] != '/' && endpoint[0] != '/' {
		// adds potential missing slash between baseURI and endpoint
		uri += baseURI + "/" + endpoint
	} else {
		uri += baseURI + endpoint
	}

	if len(requestParams) > 0 {
		uri += "?"
		for key, param := range requestParams {
			// checks for parameter pollution notation (eg. key:2)
			if key[len(key)-2] == ':' && (key[len(key)-1]-'0' >= 0 && key[len(key)-1]-'0' <= 9) {
				key = key[:len(key)-2]
			}
			uri += key + "=" + param + "&"
		}
		uri = uri[:len(uri)-1]
	}

	return uri
}

func buildBody(contentType string, bodyParams map[string]interface{}) []byte {
	var body []byte
	switch contentType {
	case "JSON":
		body = buildJSONBody(bodyParams)
	default:
		panic("Unknown content type")
	}
	return body
}

func buildJSONBody(bodyParams map[string]interface{}) []byte {
	body, err := json.Marshal(bodyParams)
	if err != nil {
		panic(err)
	}
	return body
}

func buildContentType(encodedContentType string) string {
	contentType := ""
	switch encodedContentType {
	case "JSON":
		contentType += "application/json"
	case "FORM-DATA":
		contentType += "application/x-www-form-urlencoded"
	}
	return contentType
}
