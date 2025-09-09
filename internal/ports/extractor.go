package ports
type Extractor interface {
	Extract(inputPath string) (string, error)
	ExtractFromBytes(data []byte, ext string) (string, error)
	CanHandle(ext string) bool
}
