package repo

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"

	"detector_plagio/backend/internal/config"
	"detector_plagio/backend/internal/domain"
	"detector_plagio/backend/internal/ports"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type FSUserRepo struct {
	cfg      *config.Config
	mu       sync.RWMutex
	users    map[string]*domain.User
	filePath string
}

func NewFSUserRepo(cfg *config.Config) (ports.UserRepo, error) {
	repo := &FSUserRepo{
		cfg:      cfg,
		users:    make(map[string]*domain.User),
		filePath: filepath.Join(cfg.DataRoot, "users.json"),
	}
	if err := repo.load(); err != nil {
		if os.IsNotExist(err) {
			// Create a default admin user
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
			adminUser := &domain.User{
				ID:       uuid.NewString(),
				Name:     "Admin",
				LastName: "User",
				Username: "admin",
				Password: string(hashedPassword),
				Email:    "admin@example.com",
			}
			repo.users[adminUser.ID] = adminUser
			if err := repo.save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return repo, nil
}

func (r *FSUserRepo) load() error {
	log.Println("FSUserRepo: Loading users from file")
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		log.Printf("FSUserRepo: Error reading users file %s: %v", r.filePath, err)
		return err
	}
	var users []*domain.User
	if err := json.Unmarshal(data, &users); err != nil {
		log.Printf("FSUserRepo: Error unmarshalling users data: %v", err)
		return err
	}
	r.users = make(map[string]*domain.User)
	for _, u := range users {
		r.users[u.ID] = u
	}
	log.Printf("FSUserRepo: Loaded %d users", len(r.users))
	return nil
}

func (r *FSUserRepo) save() error {
	log.Println("FSUserRepo: Saving users to file")
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*domain.User
	for _, u := range r.users {
		users = append(users, u)
	}
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		log.Printf("FSUserRepo: Error marshalling users: %v", err)
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *FSUserRepo) Create(user *domain.User) error {
	log.Println("FSUserRepo: Create method called")
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		log.Println("FSUserRepo: User with this ID already exists")
		return errors.New("user with this ID already exists")
	}
	for _, u := range r.users {
		if u.Username == user.Username {
			log.Println("FSUserRepo: Username already taken")
			return errors.New("username already taken")
		}
	}

	r.users[user.ID] = user
	log.Println("FSUserRepo: User added to map, calling save()")
	return r.save()
}

func (r *FSUserRepo) GetAll() ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []*domain.User
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

func (r *FSUserRepo) GetByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *FSUserRepo) GetByUsername(username string) (*domain.User, error) {
	log.Printf("FSUserRepo: GetByUsername called for username: %s", username)
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, u := range r.users {
		if u.Username == username {
			log.Printf("FSUserRepo: Found user with username: %s", username)
			return u, nil
		}
	}
	log.Printf("FSUserRepo: User with username %s not found", username)
	return nil, errors.New("user not found")
}

func (r *FSUserRepo) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[user.ID]; !ok {
		return errors.New("user not found")
	}
	r.users[user.ID] = user
	return r.save()
}

func (r *FSUserRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.users[id]; !ok {
		return errors.New("user not found")
	}
	delete(r.users, id)
	return r.save()
}