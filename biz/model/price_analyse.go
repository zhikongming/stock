package model

type PriceAnalyseType int

const (
	PriceAnalyseTypeAdd PriceAnalyseType = iota
	PriceAnalyseTypeDelete
)

// UpdatePriceAnalyseReq 更新量价分析股票请求
type UpdatePriceAnalyseReq struct {
	PriceAnalyseType PriceAnalyseType `json:"price_analyse_type"`
	CodeList         []string         `json:"code_list"`
}

// GetPriceAnalyseReq 获取量价分析结果请求
type GetPriceAnalyseReq struct {
}

// PriceAnalyseResult 量价分析结果
type PriceAnalyseResult struct {
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	IsSafe         string `json:"is_safe"`
	AnalysisResult string `json:"analysis_result"`
	Code           string `json:"code"`
	Name           string `json:"name"`
}

// PriceAnalyseResp 量价分析响应
type PriceAnalyseResp struct {
	Data []*PriceAnalyseResult `json:"data"`
}

type PriceAnalyseReport struct {
	EndDate string                    `json:"end_date"`
	Items   []*PriceAnalyseReportItem `json:"items"`
}

type PriceAnalyseReportItem struct {
	Name   string `json:"name"`
	IsSafe string `json:"is_safe"`
	Count  int    `json:"count"`
}

type VolumeReportItem struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	PreAmount     int64   `json:"pre_amount"`
	CurrentAmount int64   `json:"current_amount"`
	PreDate       string  `json:"pre_date"`
	CurrentDate   string  `json:"current_date"`
	Diff          float64 `json:"diff"`
	Error         error   `json:"-"`
	IndustryName  string  `json:"industry_name"`
}

// UpTrendReportItem 上升趋势报告项
// 包含股票名称、所属板块、均线金叉日期、持续天数
// 金叉：短期均线上穿长期均线，通常是买入信号
type UpTrendReportItem struct {
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	IndustryName    string  `json:"industry_name"`
	GoldCrossDate   string  `json:"gold_cross_date"`   // 均线金叉日期
	DurationDays    int     `json:"duration_days"`     // 持续天数
	LastPrice       float64 `json:"last_price"`        // 最新价格
	PriceChangeRate float64 `json:"price_change_rate"` // 价格变化率
}
