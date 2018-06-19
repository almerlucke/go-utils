package authtoken

import (
	"context"
	"net/http"
	"strings"

	"github.com/almerlucke/go-utils/server/auth/jwt"
	"github.com/almerlucke/go-utils/server/response"

	contextUtils "github.com/almerlucke/go-utils/server/context"
)

const (
	// AuthTokenKey to get auth token
	AuthTokenKey = contextUtils.Key("auth-token")
)

// Middleware middleware
type Middleware struct {
	Factory jwt.TokenDataFactory
	Secret  string
}

// New auth token middleware
func New(factory jwt.TokenDataFactory, secret string) *Middleware {
	return &Middleware{
		Factory: factory,
		Secret:  secret,
	}
}

func (ware *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	authFields := strings.Fields(authHeader)

	// Check if header contains Bearer string and token
	if len(authFields) != 2 {
		response.Unauthorized(rw, "invalid Authorization header")
		return
	}

	if authFields[0] != "Bearer" {
		response.Unauthorized(rw, "invalid Authorization header")
		return
	}

	// Unpack JWT token
	tokenData, err := jwt.UnpackToken(authFields[1], ware.Secret, ware.Factory)
	if err != nil {
		response.Unauthorized(rw, err.Error())
		return
	}

	// Add token to context
	next(rw, r.WithContext(context.WithValue(r.Context(), AuthTokenKey, tokenData)))
}

// GetAuthToken get auth token from context
func GetAuthToken(ctx context.Context) jwt.TokenData {
	return ctx.Value(AuthTokenKey).(jwt.TokenData)
}
