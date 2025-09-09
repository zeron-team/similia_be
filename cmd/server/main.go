package main

import (
	"fmt"
	"log"
	"net/http"

	"detector_plagio/backend/internal/api"
	"detector_plagio/backend/internal/config"
	"detector_plagio/backend/internal/repo"
	"detector_plagio/backend/internal/service"
	"detector_plagio/backend/internal/usecase"
	"detector_plagio/backend/internal/ports"
)

func main() {
	cfg := config.Load()
	repoFS := repo.NewFSRepo(cfg)
	userRepo, err := repo.NewFSUserRepo(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var extractors []ports.Extractor = []ports.Extractor{
		service.NewPDFToTextExtractor(),
		service.NewDocxSofficeExtractor(),
	}
	normalizer := service.NewNormalizer()
	sim := service.NewSimilarity()
	jwt := service.NewJWT(cfg.JWTSecret)

	ingest := usecase.NewIngest(cfg, repoFS, extractors, normalizer)
	compare := usecase.NewCompare(repoFS, normalizer, sim)
	auth := usecase.NewAuth(userRepo)
	user := usecase.NewUser(userRepo)

	handlers := api.NewHandlers(cfg, repoFS, userRepo, ingest, compare, auth, user, jwt)

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /login", handlers.Login)
	mux.HandleFunc("POST /documents/upload", handlers.Upload)
	mux.HandleFunc("GET /documents/{id}", handlers.GetDoc)
	mux.HandleFunc("GET /documents/ids", handlers.ListIDs)
	mux.HandleFunc("GET /documents", handlers.ListDocs)
	mux.HandleFunc("DELETE /documents/{id}", handlers.DeleteDoc)
	mux.HandleFunc("GET /folders", handlers.ListFolders)
	mux.HandleFunc("POST /compare", handlers.Compare)
	mux.HandleFunc("GET /similar/{id}", handlers.Similar)

	// Admin routes
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /users", handlers.ListUsers)
	adminMux.HandleFunc("POST /users", handlers.CreateUser)
	adminMux.HandleFunc("PUT /users/{id}", handlers.UpdateUser)
	adminMux.HandleFunc("DELETE /users/{id}", handlers.DeleteUser)
	mux.Handle("/admin/", http.StripPrefix("/admin", handlers.AuthMiddleware(adminMux)))

	addr := ":8088"
	fmt.Println("Server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func withCORS(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    // DEV: allow all origins; tighten in prod
    if origin == "" {
      origin = "*" // non-CORS (curl, etc.)
    }
    w.Header().Set("Access-Control-Allow-Origin", origin)
    w.Header().Set("Vary", "Origin")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    w.Header().Set("Access-control-allow-methods", "GET,POST,OPTIONS,DELETE,PUT")
    // some browsers expect 204 on preflight
    if r.Method == http.MethodOptions {
      w.WriteHeader(http.StatusNoContent)
      return
    }
    next.ServeHTTP(w, r)
  })
}
