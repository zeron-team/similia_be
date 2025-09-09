package ports

type MatchingSegment struct {
	TextA string `json:"textA"`
	TextB string `json:"textB"`
	Score float64 `json:"score"`
}

type JaccardResult struct {
	Intersection int     `json:"intersection"`
	Union        int     `json:"union"`
	Score        float64 `json:"score"`
}

type CosineTFIDFResult struct {
	Dot   float64 `json:"dot"`
	Na2   float64 `json:"na2"`
	Nb2   float64 `json:"nb2"`
	Score float64 `json:"score"`
}

type Similarity interface {
	Jaccard(a, b []string) JaccardResult
	CosineTFIDF(aTokens, bTokens []string) CosineTFIDFResult
	CompareSegments(textA, textB string) ([]MatchingSegment, error)
}
