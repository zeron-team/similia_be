package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"detector_plagio/backend/internal/config"
	"detector_plagio/backend/internal/ports"
	"detector_plagio/backend/internal/service"
	"detector_plagio/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type Handlers struct {
	cfg      *config.Config
	repo     ports.DocumentRepo
	userRepo ports.UserRepo
	ingest   *usecase.Ingest
	compare  *usecase.Compare
	auth     *usecase.Auth
	user     *usecase.User
	jwt      *service.JWT
}

func NewHandlers(cfg *config.Config, repo ports.DocumentRepo, userRepo ports.UserRepo, ingest *usecase.Ingest, comp *usecase.Compare, auth *usecase.Auth, user *usecase.User, jwt *service.JWT) *Handlers {
	return &Handlers{cfg: cfg, repo: repo, userRepo: userRepo, ingest: ingest, compare: comp, auth: auth, user: user, jwt: jwt}
}

func (h *Handlers) Upload(w http.ResponseWriter, r *http.Request) {
	log.Println("Upload handler called")
	if err := r.ParseMultipartForm(h.cfg.MaxUploadMB << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		id = uuid.NewString()
	}
	folder := r.FormValue("folder")
	originalFilename := r.FormValue("originalFilename")
	log.Printf("id: %s, folder: %s, originalFilename: %s", id, folder, originalFilename)
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file missing", 400)
		return
	}
	defer file.Close()
	b, _ := io.ReadAll(file)
	doc, err := h.ingest.SaveAndIndex(id, folder, originalFilename, b)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, doc)
}

func (h *Handlers) GetDoc(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := h.repo.Get(id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	writeJSON(w, d)
}

func (h *Handlers) ListDocs(w http.ResponseWriter, r *http.Request) {
	log.Println("ListDocs handler called")
	docs, err := h.repo.List()
	if err != nil {
		log.Printf("Error listing documents: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	log.Printf("Returning %d documents", len(docs))
	writeJSON(w, docs)
}

func (h *Handlers) ListIDs(w http.ResponseWriter, r *http.Request) {
	ids, err := h.repo.ListIDs()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, ids)
}

func (h *Handlers) ListFolders(w http.ResponseWriter, r *http.Request) {
	folders, err := h.repo.ListFolders()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, folders)
}

func (h *Handlers) DeleteDoc(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.repo.Delete(id); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "not found", 404)
			return
		}
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) Compare(w http.ResponseWriter, r *http.Request) {
	var p struct{ ID1, ID2 string }
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	res, err := h.compare.CompareTwo(p.ID1, p.ID2)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, res)
}

func (h *Handlers) Similar(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	topK := 10
	if q := r.URL.Query().Get("topK"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 100 {
			topK = n
		}
	}
	ids, err := h.repo.ListIDs()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type row struct {
		ID                   string `json:"id"`
		Final                int    `json:"finalPercent"`
		Near                 int    `json:"nearDuplicatePercent"`
		Topic                int    `json:"topicSimilarityPercent"`
	}
	var results []row
	for _, other := range ids {
		if other == id {
			continue
		}
		res, err := h.compare.CompareTwo(id, other)
		if err == nil {
			results = append(results, row{
				ID:    other,
				Final: int(res.Final*100 + 0.5),
				Near:  int(res.NearDuplicate*100 + 0.5),
				Topic: int(res.TopicSimilarity*100 + 0.5),
			})
		}
	}
	// sort by Final desc (simple)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Final > results[i].Final {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	if len(results) > topK {
		results = results[:topK]
	}
	writeJSON(w, results)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login handler called")
	var p struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Printf("Error decoding login request body: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}
	log.Printf("Received login data for username: %s", p.Username)

	user, err := h.auth.Login(p.Username, p.Password)
	if err != nil {
		log.Printf("Authentication failed for username %s: %v", p.Username, err)
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	log.Printf("User %s authenticated successfully", user.Username)

	token, err := h.jwt.GenerateToken(user)
	if err != nil {
		log.Printf("Error generating token for user %s: %v", user.Username, err)
		http.Error(w, "could not generate token", http.StatusInternalServerError)
		return
	}
	log.Printf("Token generated for user %s", user.Username)

	writeJSON(w, map[string]string{"token": token})
}

func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUser handler called")
	var p struct {
		Name     string `json:"name"`
		LastName string `json:"lastName"`
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), 400)
		return
	}
	log.Printf("Received user data: %+v", p)

	user, err := h.user.CreateUser(p.Name, p.LastName, p.Username, p.Password, p.Email)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	log.Printf("User created successfully: %+v", user)
	writeJSON(w, user)
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var p struct {
		Name     string `json:"name"`
		LastName string `json:"lastName"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	user, err := h.user.UpdateUser(id, p.Name, p.LastName, p.Username, p.Email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeJSON(w, user)
}

func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.user.DeleteUser(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.user.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, users)
}

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := h.jwt.ValidateToken(tokenString)
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "invalid user id in token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}