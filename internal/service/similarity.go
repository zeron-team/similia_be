package service

import (
	"log"
	"math"
	"regexp"
	"strings"

	"detector_plagio/backend/internal/ports"
)

type SimilarityService struct{}

func NewSimilarity() ports.Similarity { return &SimilarityService{} }

func (s *SimilarityService) Jaccard(a, b []string) ports.JaccardResult {
	log.Printf("Jaccard called with len(a)=%d, len(b)=%d", len(a), len(b))
	set := func(xs []string) map[string]bool {
		m := map[string]bool{}
		for _, x := range xs {
			m[x] = true
		}
		return m
	}
	A := set(a)
	B := set(b)
	if len(A) == 0 && len(B) == 0 {
		log.Println("Jaccard: both empty, returning 1.0")
		return ports.JaccardResult{Intersection: 0, Union: 0, Score: 1.0}
	}
	inter, uni := 0, 0
	seen := map[string]bool{}
	for k := range A {
		seen[k] = true
		if B[k] {
			inter++
		}
		uni++
	}
	for k := range B {
		if !seen[k] {
			uni++
		}
	}
	if uni == 0 {
		log.Println("Jaccard: union empty, returning 0")
		return ports.JaccardResult{Intersection: inter, Union: uni, Score: 0}
	}
	score := float64(inter) / float64(uni)
	log.Printf("Jaccard: intersection=%d, union=%d, score=%f", inter, uni, score)
	return ports.JaccardResult{Intersection: inter, Union: uni, Score: score}
}

func (s *SimilarityService) CosineTFIDF(aTokens, bTokens []string) ports.CosineTFIDFResult {
	log.Printf("CosineTFIDF called with len(aTokens)=%d, len(bTokens)=%d", len(aTokens), len(bTokens))

	// Term frequency for document A
	tfa := make(map[string]int)
	for _, token := range aTokens {
		tfa[token]++
	}

	// Term frequency for document B
	tfb := make(map[string]int)
	for _, token := range bTokens {
		tfb[token]++
	}

	// Document frequency (in our case, the corpus is just two documents)
	df := make(map[string]int)
	for token := range tfa {
		df[token]++
	}
	for token := range tfb {
		df[token]++
	}

	// All unique terms in both documents
	allTerms := make(map[string]bool)
	for token := range tfa {
		allTerms[token] = true
	}
	for token := range tfb {
		allTerms[token] = true
	}

	var dot, na2, nb2 float64

	for term := range allTerms {
		// TF-IDF for term in document A
		tfA := float64(tfa[term])
		idfA := 1.0 + math.Log(2.0/float64(df[term]))
		tfidfA := tfA * idfA

		// TF-IDF for term in document B
		tfB := float64(tfb[term])
		idfB := 1.0 + math.Log(2.0/float64(df[term]))
		tfidfB := tfB * idfB

		dot += tfidfA * tfidfB
		na2 += tfidfA * tfidfA
		nb2 += tfidfB * tfidfB
	}

	score := 0.0
	if na2 > 0 && nb2 > 0 {
		score = dot / (math.Sqrt(na2) * math.Sqrt(nb2))
	}

	log.Printf("CosineTFIDF: dot=%f, na2=%f, nb2=%f, score=%f", dot, na2, nb2, score)
	return ports.CosineTFIDFResult{Dot: dot, Na2: na2, Nb2: nb2, Score: score}
}

func (s *SimilarityService) CompareSegments(textA, textB string) ([]ports.MatchingSegment, error) {
	segmentsA := s.segmentText(textA)
	segmentsB := s.segmentText(textB)

	var matches []ports.MatchingSegment
	for _, segA := range segmentsA {
		for _, segB := range segmentsB {
			// Use Jaccard similarity for segment comparison
			jaccardResult := s.Jaccard(strings.Fields(strings.ToLower(segA)), strings.Fields(strings.ToLower(segB)))
			if jaccardResult.Score > 0.5 { // Threshold for considering a match
				matches = append(matches, ports.MatchingSegment{
					TextA: segA,
					TextB: segB,
					Score: jaccardResult.Score,
				})
			}
		}
	}
	return matches, nil
}

// segmentText is a basic sentence segmentation. In a real application, a more robust NLP library would be used.
func (s *SimilarityService) segmentText(text string) []string {
	// Split by common sentence-ending punctuation, followed by a space or end of string
	sentences := regexp.MustCompile(`[.!?]\s*`).Split(text, -1)
	var result []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
