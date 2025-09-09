package service

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"detector_plagio/backend/internal/ports"
)

type DocxSofficeExtractor struct{}
func NewDocxSofficeExtractor() ports.Extractor { return &DocxSofficeExtractor{} }
func (e *DocxSofficeExtractor) CanHandle(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".docx" || ext == ".txt"
}
func (e *DocxSofficeExtractor) Extract(inputPath string) (string, error) {
	if strings.HasSuffix(strings.ToLower(inputPath), ".txt") {
		b, err := os.ReadFile(inputPath); return string(b), err
	}
	tmp := os.TempDir()
	cmd := exec.Command("soffice", "--headless", "--convert-to", "txt:Text", "--outdir", tmp, inputPath)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil { return "", err }
	outPath := filepath.Join(tmp, strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))+".txt")
	defer os.Remove(outPath)
	b, err := os.ReadFile(outPath)
	return string(b), err
}

func (e *DocxSofficeExtractor) ExtractFromBytes(data []byte, ext string) (string, error) {
	log.Printf("ExtractFromBytes for docx called with data length: %d, ext: %s", len(data), ext)
	if strings.ToLower(ext) == ".txt" {
		log.Println("ExtractFromBytes: handling .txt directly")
		return string(data), nil
	}

	tmpDir := os.TempDir()
	// Create a temporary input file
	inputFile, err := os.CreateTemp(tmpDir, "soffice_input_*.docx")
	if err != nil {
		log.Printf("ExtractFromBytes: error creating temp file: %v", err)
		return "", err
	}
	defer os.Remove(inputFile.Name())
	defer inputFile.Close()

	if _, err := inputFile.Write(data); err != nil {
		log.Printf("ExtractFromBytes: error writing to temp file: %v", err)
		return "", err
	}

	// Ensure the file is written to disk before soffice tries to read it
	if err := inputFile.Sync(); err != nil {
		log.Printf("ExtractFromBytes: error syncing temp file: %v", err)
		return "", err
	}

	// soffice conversion
	cmd := exec.Command("soffice", "--headless", "--convert-to", "txt:Text", "--outdir", tmpDir, inputFile.Name())
	var errb bytes.Buffer
	cmd.Stderr = &errb
	log.Printf("ExtractFromBytes: running soffice command: %s %v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Printf("ExtractFromBytes: soffice command failed: %v, stderr: %s", err, errb.String())
		return "", errors.New(err.Error() + ": " + errb.String())
	}

	outputFileName := strings.TrimSuffix(filepath.Base(inputFile.Name()), filepath.Ext(inputFile.Name())) + ".txt"
	outPath := filepath.Join(tmpDir, outputFileName)
	defer os.Remove(outPath)

	b, err := os.ReadFile(outPath)
	if err != nil {
		log.Printf("ExtractFromBytes: error reading output file: %v", err)
		return "", err
	}
	log.Printf("ExtractFromBytes: successfully extracted text with length %d", len(b))
	return string(b), nil
}
