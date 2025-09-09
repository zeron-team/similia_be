package service

import (
	"bytes"
	"log"
	"os/exec"
	"strings"

	"detector_plagio/backend/internal/ports"
)

type PDFToTextExtractor struct{}
func NewPDFToTextExtractor() ports.Extractor { return &PDFToTextExtractor{} }
func (e *PDFToTextExtractor) CanHandle(ext string) bool { return strings.EqualFold(ext, ".pdf") }
func (e *PDFToTextExtractor) Extract(inputPath string) (string, error) {
	cmd := exec.Command("pdftotext", "-layout", inputPath, "-")
	var out, errb bytes.Buffer
	cmd.Stdout = &out; cmd.Stderr = &errb
	if err := cmd.Run(); err != nil { return "", err }
	return strings.TrimSpace(out.String()), nil
}

func (e *PDFToTextExtractor) ExtractFromBytes(data []byte, ext string) (string, error) {
	log.Printf("ExtractFromBytes for pdf called with data length: %d, ext: %s", len(data), ext)
	cmd := exec.Command("pdftotext", "-layout", "-", "-")
	cmd.Stdin = bytes.NewReader(data)
	var out, errb bytes.Buffer
	cmd.Stdout = &out; cmd.Stderr = &errb
	log.Printf("ExtractFromBytes: running pdftotext command: %s %v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		log.Printf("ExtractFromBytes: pdftotext command failed: %v, stderr: %s", err, errb.String())
		return "", err
	}
	log.Printf("ExtractFromBytes: successfully extracted text with length %d", len(out.String()))
	return strings.TrimSpace(out.String()), nil
}
