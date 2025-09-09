package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DataRoot      string
	MaxUploadMB   int64
	AllowedExtMap map[string]bool
	JWTSecret     string
}

func Load() *Config {
	root := os.Getenv("DOCSIM_DATA_ROOT")
	if root == "" {
		root = "./data"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "a-very-secret-key" // a default for development
	}

	cfg := &Config{
		DataRoot:    root,
		MaxUploadMB: 50,
		AllowedExtMap: map[string]bool{
			".pdf":  true,
			".docx": true,
			".txt":  true,
		},
		JWTSecret: jwtSecret,
	}
	_ = os.MkdirAll(filepath.Join(root, "docs"), 0755)
	_ = os.MkdirAll(filepath.Join(root, "texts"), 0755)
	_ = os.MkdirAll(filepath.Join(root, "index"), 0755)
	return cfg
}

func (c *Config) DocsPath() string  { return filepath.Join(c.DataRoot, "docs") }
func (c *Config) TextsPath() string { return filepath.Join(c.DataRoot, "texts") }
