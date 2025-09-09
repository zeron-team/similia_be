package ports

import "detector_plagio/backend/internal/domain"

// UserRepo is the interface for user persistence.
type UserRepo interface {
	Create(user *domain.User) error
	GetAll() ([]*domain.User, error)
	GetByID(id string) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id string) error
}
