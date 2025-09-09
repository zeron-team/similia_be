package ports
type Normalizer interface {
	Normalize(s string) string
	Tokenize(s string) []string
	Shingles(tokens []string, k int) []string
}
