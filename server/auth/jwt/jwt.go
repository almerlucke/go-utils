package jwt

import (
	jwt "github.com/dgrijalva/jwt-go"
)

// TokenData token data interface
type TokenData interface {
	// GetClaims data to claims
	GetClaims() jwt.MapClaims

	// SetClaims claims to data
	SetClaims(claims jwt.MapClaims) error
}

// TokenDataFactory creates new token data instances
type TokenDataFactory interface {
	New() TokenData
}

// GenerateToken generate JWT token
func GenerateToken(signingSecret string, issuedAt int64, expiresAfter int64, tokenData TokenData) (string, error) {
	// Get claims of token data object
	otherClaims := tokenData.GetClaims()

	// Always populate issued at and expires
	claims := jwt.MapClaims{
		"iat": issuedAt,
		"exp": expiresAfter,
	}

	for key, val := range otherClaims {
		claims[key] = val
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(signingSecret))
}

// UnpackToken validate and unpack JWT token data
func UnpackToken(signedString string, signingSecret string, factory TokenDataFactory) (TokenData, error) {
	// Generate new token data
	tokenData := factory.New()

	// Parse token
	token, err := jwt.Parse(signedString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.NewValidationError("Invalid JWT token", 0)
		}

		return []byte(signingSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check claims and if token is valid
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.NewValidationError("Invalid JWT token", 0)
	}

	// Set claims from token
	err = tokenData.SetClaims(claims)
	if err != nil {
		return nil, err
	}

	return tokenData, nil
}
