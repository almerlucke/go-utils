package handles

import (
	"net/http"
	"strings"

	"github.com/almerlucke/go-utils/server/auth/basic"
	"github.com/almerlucke/go-utils/server/auth/jwt"
	"github.com/almerlucke/go-utils/server/request/localization"
	"github.com/almerlucke/kobogo/server/response"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/text/language"
)

// Handle for routes without authentication
type Handle func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params, loc *localization.Localization)

// JWTAuthHandle for routes which need JWT authentication
type JWTAuthHandle func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params, tokenData jwt.TokenData, loc *localization.Localization)

// Wrap with localization in httprouter handle
func Wrap(languageMatcher language.Matcher, handle Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		handle(rw, r, pm, localization.GetLocalizationForRequest(r, languageMatcher))
	}
}

// JWTAuthWrap wraps another handle and perform JWT authentication before calling the given handle
func JWTAuthWrap(signingSecret string, languageMatcher language.Matcher, factory jwt.TokenDataFactory, handle JWTAuthHandle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params) {

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		authFields := strings.Fields(authHeader)

		// Check if header contains Bearer string and token
		if len(authFields) != 2 {
			response.Unauthorized(rw, "not a valid Authorization header")
			return
		}

		if authFields[0] != "Bearer" {
			response.Unauthorized(rw, "not a valid Authorization header")
			return
		}

		// Unpack JWT token
		tokenData, err := jwt.UnpackToken(authFields[1], signingSecret, factory)
		if err != nil {
			response.Unauthorized(rw, err.Error())
			return
		}

		// Call handle
		handle(rw, r, pm, tokenData, localization.GetLocalizationForRequest(r, languageMatcher))
	}
}

// BasicAuthWrap wraps another handle and performs basic authentication before calling the given handle
func BasicAuthWrap(username string, password string, languageMatcher language.Matcher, handle Handle) httprouter.Handle {
	return func(rw http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		if !basic.ValidateBasicAuthHeader(r.Header.Get("Authorization"), username, password) {
			rw.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
			response.Unauthorized(rw, "invalid username and password")
			return
		}

		handle(rw, r, pm, localization.GetLocalizationForRequest(r, languageMatcher))
	}
}
