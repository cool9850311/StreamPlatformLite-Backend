package util

import (
	"golang.org/x/crypto/bcrypt"
)

type BcryptLibrary struct{}

func NewBcryptLibrary() *BcryptLibrary {
	return &BcryptLibrary{}
}

func (b *BcryptLibrary) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (b *BcryptLibrary) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
