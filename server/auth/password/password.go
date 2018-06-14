package password

import "golang.org/x/crypto/bcrypt"

// GetPasswordHash creates a bcrypt password hash
func GetPasswordHash(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashBytes), nil
}

// CheckHashAndPassword checks a hash against a password
func CheckHashAndPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
