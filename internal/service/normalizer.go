package service

import (
	"regexp"
	"strings"

	"detector_plagio/backend/internal/ports"
)

var nonLetter = regexp.MustCompile(`[^a-zA-Z0-9áéíóúÁÉÍÓÚñÑüÜ]`)

type SimpleNormalizer struct{}
func NewNormalizer() ports.Normalizer { return &SimpleNormalizer{} }

func (n *SimpleNormalizer) Normalize(s string) string {
	s = strings.ToLower(s)
	s = nonLetter.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}
func (n *SimpleNormalizer) Tokenize(s string) []string {
	toks := strings.Fields(n.Normalize(s))
	stop := map[string]bool{
		"the": true,"and": true,"of": true,"to": true,"a": true,"in": true,"for": true,"is": true,
		"de": true,"la": true,"el": true,"y": true,"los": true,"las": true,"un": true,"una": true,"en": true,
	}
	out := make([]string, 0, len(toks))
	for _, t := range toks { if !stop[t] { out = append(out, t) } }
	return out
}
func (n *SimpleNormalizer) Shingles(tokens []string, k int) []string {
	if k <= 1 || len(tokens) < k { return tokens }
	out := make([]string, 0, len(tokens)-k+1)
	for i := 0; i <= len(tokens)-k; i++ { out = append(out, strings.Join(tokens[i:i+k], " ")) }
	return out
}
