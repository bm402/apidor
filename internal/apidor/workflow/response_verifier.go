package workflow

import (
	"net/http"
	"strconv"
)

func verifyResponseExpectedOK(response *http.Response) string {
	defer response.Body.Close()
	result := strconv.Itoa(response.StatusCode) + " "
	if response.StatusCode/100 != 2 {
		result += "Unexpected status code, expecting 2xx"
	} else {
		result += "OK"
	}
	return result
}

func verifyResponseExpectedUnauthorised(response *http.Response) string {
	defer response.Body.Close()
	result := strconv.Itoa(response.StatusCode) + " "
	if response.StatusCode/100 != 4 {
		result += "Unexpected status code, expecting 4xx"
	} else {
		result += "OK"
	}
	return result
}
