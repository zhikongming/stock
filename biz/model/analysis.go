package model

type SuggestOperationPriority int
type MaType string

const (
	HighSuggestOperationPriority   SuggestOperationPriority = 1
	MediumSuggestOperationPriority SuggestOperationPriority = 2
	LowSuggestOperationPriority    SuggestOperationPriority = 3
)

// -------------- macd 分析结果 -----------------
func (p SuggestOperationPriority) ToInt() int {
	return int(p)
}

type MacdAnalyzeResult struct {
	IsBuyPoint bool
	Length     int
	Reason     string
	Priority   SuggestOperationPriority
}

// -------------- ma线 -----------------
