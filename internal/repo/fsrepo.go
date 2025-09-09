package repo

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"detector_plagio/backend/internal/config"
	"detector_plagio/backend/internal/domain"
	"detector_plagio/backend/internal/ports"
)

var idRe = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

type FSRepo struct{ cfg *config.Config }

func NewFSRepo(cfg *config.Config) ports.DocumentRepo { return &FSRepo{cfg: cfg} }

func (r *FSRepo) Save(doc domain.Document, data []byte) error {
	log.Printf("Saving document: %+v", doc)
	if !idRe.MatchString(doc.ID) {
		return errors.New("invalid id")
	}
	ext := strings.ToLower(doc.Ext)
	if !r.cfg.AllowedExtMap[ext] {
		return errors.New("unsupported extension")
	}
	dst := filepath.Join(r.cfg.DocsPath(), doc.ID+ext)
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return err
	}
	// lightweight metadata (optional)
	doc.UpdatedAt = time.Now().Format(time.RFC3339)
	meta := filepath.Join(r.cfg.DocsPath(), doc.ID+".json")
	b, _ := json.MarshalIndent(doc, "", "  ")
	_ = os.WriteFile(meta, b, 0644)
	log.Printf("Saved document %s to %s and %s", doc.ID, dst, meta)
	return nil
}

func (r *FSRepo) Get(id string) (domain.Document, error) {
	var d domain.Document
	if !idRe.MatchString(id) {
		return d, errors.New("invalid id")
	}

	metaPath := filepath.Join(r.cfg.DocsPath(), id+".json")
	if meta, err := os.ReadFile(metaPath); err == nil {
		if err := json.Unmarshal(meta, &d); err == nil {
			return d, nil
		}
	}

	return d, os.ErrNotExist
}

// List returns all discovered documents (based on files present with allowed extensions).
func (r *FSRepo) List() ([]domain.Document, error) {
	log.Println("List function called")
	ents, err := os.ReadDir(r.cfg.DocsPath())
	if err != nil {
		return nil, err
	}
	docs := []domain.Document{}
	seen := map[string]bool{}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if !r.cfg.AllowedExtMap[ext] {
			continue
		}
		id := strings.TrimSuffix(name, ext)
		if !idRe.MatchString(id) || seen[id] {
			continue
		}
		p := filepath.Join(r.cfg.DocsPath(), name)
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			// Check for metadata file
			metaPath := filepath.Join(r.cfg.DocsPath(), id+".json")
			var docData domain.Document
			if meta, err := os.ReadFile(metaPath); err == nil {
				if err := json.Unmarshal(meta, &docData); err != nil {
					log.Printf("Error unmarshaling metadata for %s: %v", id, err)
					continue // Skip this document if metadata is unreadable
				}
			} else {
				log.Printf("Metadata file not found for %s: %v", id, err)
				// If no metadata, use basic info from file system
				docData = domain.Document{
					ID:       id,
					Filename: name,
					Size:     info.Size(),
					Ext:      ext,
				}
			}

			docs = append(docs, docData)
			seen[id] = true
		}
	}
	log.Printf("Listed %d documents", len(docs))
	return docs, nil
}

func (r *FSRepo) ListIDs() ([]string, error) {
	ents, err := os.ReadDir(r.cfg.DocsPath())
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if r.cfg.AllowedExtMap[ext] {
			id := strings.TrimSuffix(name, ext)
			if idRe.MatchString(id) {
				seen[id] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}

// ListFolders scans sidecar JSON metadata files for an optional "folder" field
// and returns the unique set of folder names, if present.
func (r *FSRepo) ListFolders() ([]string, error) {
	ents, err := os.ReadDir(r.cfg.DocsPath())
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		p := filepath.Join(r.cfg.DocsPath(), name)
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		// parse generically so we don't depend on struct fields
		var meta map[string]any
		if err := json.Unmarshal(b, &meta); err == nil {
			if v, ok := meta["folder"].(string); ok && v != "" {
				seen[v] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for f := range seen {
		out = append(out, f)
	}
	return out, nil
}

func (r *FSRepo) PathFor(id string) (string, string) {
	raw := ""
	for ext := range r.cfg.AllowedExtMap {
		p := filepath.Join(r.cfg.DocsPath(), id+ext)
		if _, err := os.Stat(p); err == nil {
			raw = p
			break
		}
	}
	return raw, filepath.Join(r.cfg.TextsPath(), id+".txt")
}

func (r *FSRepo) Delete(id string) error {
	if !idRe.MatchString(id) {
		return errors.New("invalid id")
	}
	rawPath, txtPath := r.PathFor(id)
	if rawPath == "" {
		return os.ErrNotExist
	}
	if err := os.Remove(rawPath); err != nil {
		return err
	}
	if err := os.Remove(txtPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	metaPath := filepath.Join(r.cfg.DocsPath(), id+".json")
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
