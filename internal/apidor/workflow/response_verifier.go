package workflow

import (
	"net/http"
)

func verifyResponseExpectedOK(response *http.Response) (int, string) {
	defer response.Body.Close()
	status := response.StatusCode
	var result string

	if response.StatusCode/100 != 2 {
		result = "Unexpected status code, expecting 2xx"
	} else {
		result = "OK"
	}

	return status, result
}

func verifyResponseExpectedUnauthorised(response *http.Response) (int, string) {
	defer response.Body.Close()
	status := response.StatusCode
	var result string

	if response.StatusCode/100 != 4 {
		result = "Unexpected status code, expecting 4xx"
	} else {
		result = "OK"
	}

	return status, result
}
