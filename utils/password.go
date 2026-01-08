package utils

import "golang.org/x/crypto/bcrypt"

// DefaultBcryptCost is the cost factor for bcrypt hashing
// 10 is recommended for production (takes ~100ms)
// 14 is very secure but slow (~1.5s)
const DefaultBcryptCost = 10

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
