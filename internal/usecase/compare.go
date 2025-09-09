package usecase

import (
	"log"

	"detector_plagio/backend/internal/ports"
)

type Compare struct {
	repo ports.DocumentRepo
	norm ports.Normalizer
	sim  ports.Similarity
}
type CompareResult struct {
	Doc1TextContentLength int                       `json:"doc1TextContentLength"`
	Doc2TextContentLength int                       `json:"doc2TextContentLength"`
	Doc1TokensLength      int                       `json:"doc1TokensLength"`
	Doc2TokensLength      int                       `json:"doc2TokensLength"`
	Doc1ShinglesLength    int                       `json:"doc1ShinglesLength"`
	Doc2ShinglesLength    int                       `json:"doc2ShinglesLength"`
	Jaccard               ports.JaccardResult       `json:"jaccard"`
	Cosine                ports.CosineTFIDFResult   `json:"cosine"`
	NearDuplicate         float64                   `json:"nearDuplicate"`
	TopicSimilarity       float64                   `json:"topicSimilarity"`
	Final                 float64                   `json:"final"`
	MatchingSegments      []ports.MatchingSegment `json:"matchingSegments"`
}

func NewCompare(repo ports.DocumentRepo, n ports.Normalizer, s ports.Similarity) *Compare {
	return &Compare{repo: repo, norm: n, sim: s}
}

func (u *Compare) CompareTwo(id1, id2 string) (CompareResult, error) {
	log.Printf("Comparing %s and %s", id1, id2)
	doc1, err := u.repo.Get(id1)
	if err != nil {
		return CompareResult{}, err
	}
	doc2, err := u.repo.Get(id2)
	if err != nil {
		return CompareResult{}, err
	}

	log.Printf("Doc1 TextContent length: %d, Doc2 TextContent length: %d", len(doc1.TextContent), len(doc2.TextContent))
	tok1 := u.norm.Tokenize(doc1.TextContent)
	tok2 := u.norm.Tokenize(doc2.TextContent)
	log.Printf("Doc1 tokens length: %d, Doc2 tokens length: %d", len(tok1), len(tok2))
	sh1 := u.norm.Shingles(tok1, 5)
	sh2 := u.norm.Shingles(tok2, 5)
	log.Printf("Doc1 shingles length: %d, Doc2 shingles length: %d", len(sh1), len(sh2))

	jaccardResult := u.sim.Jaccard(sh1, sh2)
	cosineResult := u.sim.CosineTFIDF(tok1, tok2)
	near := jaccardResult.Score
	topic := cosineResult.Score
	final := 0.6*near + 0.4*topic
	log.Printf("Near: %f, Topic: %f, Final: %f", near, topic, final)

	matchingSegments, err := u.sim.CompareSegments(doc1.TextContent, doc2.TextContent)
	if err != nil {
		return CompareResult{}, err
	}
	log.Printf("Found %d matching segments", len(matchingSegments))

	return CompareResult{
		Doc1TextContentLength: len(doc1.TextContent),
		Doc2TextContentLength: len(doc2.TextContent),
		Doc1TokensLength:      len(tok1),
		Doc2TokensLength:      len(tok2),
		Doc1ShinglesLength:    len(sh1),
		Doc2ShinglesLength:    len(sh2),
		Jaccard:               jaccardResult,
		Cosine:                cosineResult,
		NearDuplicate:         near,
		TopicSimilarity:       topic,
		Final:                 final,
		MatchingSegments:      matchingSegments,
	}, nil
}
