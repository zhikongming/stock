package model

type AnalyzeReport struct{}

type CodeScore struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
}

type CodeScoreSorter []*CodeScore

func (s CodeScoreSorter) Len() int {
	return len(s)
}
func (s CodeScoreSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s CodeScoreSorter) Less(i, j int) bool {
	return s[i].Value > s[j].Value
}

type SimpleScoreResult struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

type ScoreResult struct {
	Code           string                            `json:"code"`
	Name           string                            `json:"name"`
	Price          float64                           `json:"price"`
	ChangeScore    float64                           `json:"change_score"`
	Slope5Score    float64                           `json:"slope5_score"`
	Slope10Score   float64                           `json:"slope10_score"`
	Slope20Score   float64                           `json:"slope20_score"`
	Score5gt10     float64                           `json:"score5gt10"`
	Score10gt20    float64                           `json:"score10gt20"`
	Score20gt30    float64                           `json:"score20gt30"`
	NewHighScore   float64                           `json:"new_high_score"`
	VolumeScore    float64                           `json:"volume_score"`
	RPS20Score     float64                           `json:"rps20_score"`
	RPS5Score      float64                           `json:"rps5_score"`
	Score          float64                           `json:"score"`
	MaxStockCode   string                            `json:"max_stock_code"`
	MaxStockName   string                            `json:"max_stock_name"`
	MaxStockChange float64                           `json:"max_stock_change"`
	ThirdBuyPoint  []*FilterThirdBuyCodePeriodResult `json:"third_buy_code"`
}

type ScoreResultSorter []*ScoreResult

func (s ScoreResultSorter) Len() int {
	return len(s)
}
func (s ScoreResultSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ScoreResultSorter) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

type ScoreResultDiff struct {
	ScoreDiff string `json:"score_diff"`
	OrderDiff string `json:"order_diff"`
}
