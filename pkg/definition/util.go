package definition

import (
	"strings"
)

func buildHighPrivilegedAuthHeader(authDetails AuthDetails) (string, string) {
	headerValue := ""
	if len(authDetails.headerValuePrefix) > 0 {
		headerValue += strings.TrimSpace(authDetails.headerValuePrefix) + " "
	}
	headerValue += strings.TrimSpace(authDetails.high)
	return authDetails.headerName, headerValue
}
