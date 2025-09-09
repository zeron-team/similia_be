package domain

// User represents a user of the application.
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastName"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
