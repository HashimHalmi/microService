package microServerMainFiles

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"microService/pkg/email"
)

type UserCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"` // This should be hashed before storage
}

func RegisterUser(user UserCredentials) error {
	password := GenerateRandomPassword()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err // return error if hashing failed
	}
	user.Password = string(hashedPassword) // Save hashed password
	if err := SaveUser(user); err != nil {
		return err
	}
	return email.SendEmail(user.Email, "Your Password", "Your password is: "+password)
}

func AuthenticateUser(user UserCredentials) error {
	storedUser, err := GetUserByEmail(user.Email)
	if err != nil {
		return err
	}
	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		return errors.New("authentication failed")
	}
	return nil
}

func GenerateRandomPassword() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
