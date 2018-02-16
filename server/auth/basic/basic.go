package basic

import (
	"encoding/base64"
	"strings"
)

// ValidateBasicAuthHeader validate a basic auth header string
func ValidateBasicAuthHeader(header string, user string, password string) bool {
	// Split authorization header.
	components := strings.SplitN(header, " ", 2)
	if len(components) != 2 || components[0] != "Basic" {
		return false
	}

	// Decode credential.
	credential, err := base64.StdEncoding.DecodeString(components[1])
	if err != nil {
		return false
	}

	// Split credential into userid, password.
	pair := strings.SplitN(string(credential), ":", 2)
	if len(pair) != 2 {
		return false
	}

	// Compare user and password
	return pair[0] == user && pair[1] == password
}
