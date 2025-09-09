package usecase

import (
	"log" // Add log import
	"detector_plagio/backend/internal/domain"
	"detector_plagio/backend/internal/ports"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	userRepo ports.UserRepo
}

func NewUser(userRepo ports.UserRepo) *User {
	return &User{userRepo: userRepo}
}

func (uc *User) CreateUser(name, lastName, username, password, email string) (*domain.User, error) {
	log.Println("User usecase: CreateUser called")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("User usecase: Error hashing password: %v", err)
		return nil, err
	}
	log.Println("User usecase: Password hashed")

	user := &domain.User{
		ID:       uuid.NewString(),
		Name:     name,
		LastName: lastName,
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}
	log.Printf("User usecase: User object created: %+v", user)

	err = uc.userRepo.Create(user)
	if err != nil {
		log.Printf("User usecase: Error creating user in repo: %v", err)
		return nil, err
	}
	log.Println("User usecase: User created in repo")

	return user, nil
}

func (uc *User) UpdateUser(id, name, lastName, username, email string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	user.Name = name
	user.LastName = lastName
	user.Username = username
	user.Email = email

	err = uc.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *User) DeleteUser(id string) error {
	return uc.userRepo.Delete(id)
}

func (uc *User) GetAllUsers() ([]*domain.User, error) {
	return uc.userRepo.GetAll()
}
