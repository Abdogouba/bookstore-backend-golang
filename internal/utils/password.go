package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword converts plain password into bcrypt hash.
func HashPassword(password string) (string, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	return string(hashedPassword), err
}

// CheckPassword compares plain password with hash.
func CheckPassword(
	password string,
	hash string,
) error {

	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}