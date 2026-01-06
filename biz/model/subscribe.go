package model

type StrategyType int
type PriceChangeType int

const (
	StrategyTypeIndustryRateChange StrategyType = 1
	StrategyTypeStockRateChange    StrategyType = 2
	StrategyTypeStockPriceChange   StrategyType = 3

	PriceChangeTypeGreater PriceChangeType = 1
	PriceChangeTypeLess    PriceChangeType = 2
)

type AddSubscribeStrategyReq struct {
	StrategyType    StrategyType    `json:"strategy_type"`
	IndustryCode    string          `json:"industry_code"`
	Days            int             `json:"days"`
	PriceChange     float64         `json:"price_change"`
	PriceChangeType PriceChangeType `json:"price_change_type"`
	RateChange      float64         `json:"rate_change"`
	StockCode       string          `json:"stock_code"`
}

type GetSubscribeStrategyReq struct {
	ID int `json:"id"`
}

type DeleteSubscribeStrategyReq struct {
	ID int `json:"id"`
}

type SubscribeStrategyResult struct {
	ID             int    `json:"id"`
	DateTime       string `json:"date_time"`
	StrategyType   string `json:"strategy_type"`
	Code           string `json:"code"`
	Strategy       string `json:"strategy"`
	Result         bool   `json:"result"`
	StrategyDetail string `json:"strategy_detail"`
	LastDate       string `json:"last_date"`
}

func (s PriceChangeType) String() string {
	switch s {
	case PriceChangeTypeGreater:
		return "大于"
	case PriceChangeTypeLess:
		return "小于"
	default:
		return ""
	}
}

func (s StrategyType) String() string {
	switch s {
	case StrategyTypeIndustryRateChange:
		return "板块波动率"
	case StrategyTypeStockRateChange:
		return "个股波动率"
	case StrategyTypeStockPriceChange:
		return "个股价格变动"
	default:
		return ""
	}
}
