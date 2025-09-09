package usecase

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"detector_plagio/backend/internal/config"
	"detector_plagio/backend/internal/domain"
	"detector_plagio/backend/internal/ports"
)

type Ingest struct {
	cfg        *config.Config
	repo       ports.DocumentRepo
	extractors []ports.Extractor
	norm       ports.Normalizer
}

func NewIngest(cfg *config.Config, repo ports.DocumentRepo, ex []ports.Extractor, n ports.Normalizer) *Ingest {
	return &Ingest{cfg: cfg, repo: repo, extractors: ex, norm: n}
}

func (u *Ingest) SaveAndIndex(id, folder, originalFilename string, data []byte) (domain.Document, error) {
	ext := strings.ToLower(filepath.Ext(originalFilename))
	doc := domain.Document{ID: id, Folder: folder, Filename: id + ext, OriginalFilename: originalFilename, Size: int64(len(data)), Ext: ext}
	var ex ports.Extractor
	for _, e := range u.extractors { if e.CanHandle(ext) { ex = e; break } }
	if ex == nil { return doc, errors.New("no extractor for " + ext) }
	// Extract text directly from the provided data (file content)
	text, err := ex.ExtractFromBytes(data, ext); if err != nil { return doc, err }
	doc.TextContent = u.norm.Normalize(text)
	// Save the document *after* TextContent is populated
	if err := u.repo.Save(doc, data); err != nil { return doc, err }
	_, txtPath := u.repo.PathFor(id)
	// Write the extracted text to a file
	if err := os.WriteFile(txtPath, []byte(doc.TextContent), 0644); err != nil { return doc, err }
	return doc, nil
}
