package workflow

import (
	"io/ioutil"
	"net/http"
	"strings"
)

type bannedWordCheckResult int

const (
	found bannedWordCheckResult = iota
	foundInError
	notFound
)

func verifyResponseExpectedOK(response *http.Response, bannedResponseWords []string) (int, string) {
	status := response.StatusCode
	var result string

	if response.StatusCode/100 != 2 {
		result = "Unexpected status code, expecting 2xx"
	} else {
		result = "OK"
	}

	return status, result
}

func verifyResponseExpectedUnauthorised(response *http.Response, bannedResponseWords []string) (int, string) {
	status := response.StatusCode
	var result string

	if response.StatusCode/100 != 4 {
		isFound := checkResponseBodyForBannedWords(response, bannedResponseWords)
		switch isFound {
		case found:
			result = "Unexpected status code and high privileged data found in response"
		case foundInError:
			result = "OK (unexpected status code and high privileged data found in probable error message)"
		case notFound:
			result = "OK (unexpected status code)"
		}
	} else {
		isFound := checkResponseBodyForBannedWords(response, bannedResponseWords)
		switch isFound {
		case found:
			result = "High privileged data found in response"
		case foundInError:
			result = "OK (high privileged data found in probable error message)"
		case notFound:
			result = "OK"
		}
	}

	return status, result
}

func checkResponseBodyForBannedWords(response *http.Response, bannedResponseWords []string) bannedWordCheckResult {
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return notFound
	}

	bodyStr := strings.ToLower(string(bodyBytes))
	for _, bannedWord := range bannedResponseWords {
		if strings.Contains(bodyStr, strings.ToLower(bannedWord)) {
			if strings.Contains(bodyStr, "error") {
				return foundInError
			}
			return found
		}
	}

	return notFound
}
