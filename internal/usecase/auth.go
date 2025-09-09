package usecase

import (
	"errors"
	"log" // Add log import
	"detector_plagio/backend/internal/domain"
	"detector_plagio/backend/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	userRepo ports.UserRepo
}

func NewAuth(userRepo ports.UserRepo) *Auth {
	return &Auth{userRepo: userRepo}
}

func (uc *Auth) Login(username, password string) (*domain.User, error) {
	log.Printf("Auth usecase: Login called for username: %s", username)
	user, err := uc.userRepo.GetByUsername(username)
	if err != nil {
		log.Printf("Auth usecase: User %s not found: %v", username, err)
		return nil, errors.New("invalid credentials")
	}
	log.Printf("Auth usecase: User %s found", username)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("Auth usecase: Password mismatch for user %s: %v", username, err)
		return nil, errors.New("invalid credentials")
	}
	log.Printf("Auth usecase: Password matched for user %s", username)

	return user, nil
}
