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

type CodePrice struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}

type ThirdBuyPeriod struct {
	StartDate  string  `json:"startDate"`
	StartPrice float64 `json:"startPrice"`
	EndDate    string  `json:"endDate"`
	EndPrice   float64 `json:"endPrice"`
	Rate       float64 `json:"rate"`
}

type FinalThirdBuyPeriod struct {
	StartDate  string  `json:"startDate"`
	StartPrice float64 `json:"startPrice"`
	EndPrice   float64 `json:"endPrice"`
	Rate       float64 `json:"rate"`
}

type ThirdBuyCodePeriod struct {
	UpPeriod    *ThirdBuyPeriod      `json:"upPeriod"`
	DownPeriod  *ThirdBuyPeriod      `json:"downPeriod"`
	ReupPeriod  *ThirdBuyPeriod      `json:"reupPeriod"`
	FinalPeriod *FinalThirdBuyPeriod `json:"finalPeriod"`
}

func (p *ThirdBuyCodePeriod) ValidFilter(thresholdUp float64, thresholdPullback float64, thresholdDeviation float64, thresholdProfit float64) bool {
	if p == nil {
		return false
	}
	if thresholdUp > 0 && p.UpPeriod.Rate < thresholdUp {
		return false
	}
	if thresholdPullback > 0 && p.DownPeriod.Rate < thresholdPullback {
		return false
	}
	if thresholdDeviation > 0 && p.ReupPeriod.Rate > thresholdDeviation {
		return false
	}
	if thresholdProfit > 0 && p.FinalPeriod.Rate < thresholdProfit {
		return false
	}
	return true
}

type FilterThirdBuyCodePeriodResp struct {
	Total int                               `json:"total"`
	Data  []*FilterThirdBuyCodePeriodResult `json:"data"`
}

type FilterThirdBuyCodePeriodResult struct {
	ThirdBuyCodePeriod
	Code         string `json:"code"`
	Name         string `json:"name"`
	IndustryName string `json:"industryName"`
}

type SorterFilterThirdBuyCodePeriodResult []*FilterThirdBuyCodePeriodResult

func (s SorterFilterThirdBuyCodePeriodResult) Len() int {
	return len(s)
}

func (s SorterFilterThirdBuyCodePeriodResult) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SorterFilterThirdBuyCodePeriodResult) Less(i, j int) bool {
	if s[i].FinalPeriod.Rate > s[j].FinalPeriod.Rate {
		return true
	}
	return false
}
