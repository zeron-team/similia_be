package ports

import "detector_plagio/backend/internal/domain"

type DocumentRepo interface {
	Save(doc domain.Document, data []byte) error
	Get(id string) (domain.Document, error)
	List() ([]domain.Document, error)
	ListIDs() ([]string, error)
	ListFolders() ([]string, error)
	PathFor(id string) (rawPath string, txtPath string)
	Delete(id string) error
}
