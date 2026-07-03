package password

import "golang.org/x/crypto/bcrypt"

// Hash returns a bcrypt hash of the plain-text password.
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare checks whether password matches hash.
func Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
