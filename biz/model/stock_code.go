package model

import (
	"sort"
	"strings"
)

type StockStrategy string
type StockSuggestOperation string
type StockMaType string
type BollingPosition string
type KLineType int

const (
	StockStrategyMa      StockStrategy = "ma"
	StockStrategyBolling StockStrategy = "bolling"
	StockStrategyMacd    StockStrategy = "macd"
	StockStrategyKdj     StockStrategy = "kdj"

	StockSuggestOperationBuy  StockSuggestOperation = "buy"
	StockSuggestOperationSell StockSuggestOperation = "sell"
	StockSuggestOperationHold StockSuggestOperation = "hold"
	StockSuggestOperationNone StockSuggestOperation = "none"

	StockMaType5  StockMaType = "ma5"
	StockMaType10 StockMaType = "ma10"
	StockMaType20 StockMaType = "ma20"
	StockMaType30 StockMaType = "ma30"
	StockMaType60 StockMaType = "ma60"

	BollingCmpPercent                   = 0.33
	BollingPositionUp   BollingPosition = "up"
	BollingPositionDown BollingPosition = "down"
	BollingPositionMid  BollingPosition = "mid"

	KdjOversold   = 20
	KdjOverbought = 80

	KLineTypeDay   KLineType = 0
	KLineType30Min KLineType = 1
)

type SyncStockCodeReq struct {
	Code         string `json:"code"`
	BusinessType int    `json:"business_type"`
}

type SyncStockIndustryReq struct {
}

type SyncFundFlowReq struct {
}

type AnalyzeStockCodeReq struct {
	Code     string        `json:"code"`
	Date     string        `json:"date,omitempty"`
	Strategy StockStrategy `json:"strategy"`
}

type MacdFilter struct {
	MaxLastDif float64 `json:"max_last_dif"`
	MaxLastDea float64 `json:"max_last_dea"`
	MinLength  int     `json:"min_length"`
}

type MaFilter struct {
	Ma5Position  int `json:"ma5_position"`
	Ma10Position int `json:"ma10_position"`
	Ma20Position int `json:"ma20_position"`
	Ma30Position int `json:"ma30_position"`
	Ma60Position int `json:"ma60_position"`
}

type BollingFilter struct {
	BollingPosition BollingPosition `json:"bolling_position"`
}

type KdjFilter struct {
	MaxLastK float64 `json:"max_last_k"`
	MaxLastD float64 `json:"max_last_d"`
	MaxLastJ float64 `json:"max_last_j"`
}

type FilterStockCodeReq struct {
	Date          string         `json:"date,omitempty"`
	MacdFilter    *MacdFilter    `json:"macd_filter,omitempty"`
	MaFilter      *MaFilter      `json:"ma_filter,omitempty"`
	BollingFilter *BollingFilter `json:"bolling_filter,omitempty"`
	KdjFilter     *KdjFilter     `json:"kdj_filter,omitempty"`
}

type FilterStockCodeItem struct {
	Code        string                                  `json:"code"`
	CompanyName string                                  `json:"company_name"`
	Result      map[StockStrategy]*AnalyzeStockCodeResp `json:"result"`
	LastDate    string                                  `json:"last_date"`
}

type MaOrderData struct {
	MaType  StockMaType `json:"ma_type"`
	MaPrice float64     `json:"ma_price"`
}

type MacdValue struct {
	LastDif float64 `json:"last_dif"`
	LastDea float64 `json:"last_dea"`
	Length  int     `json:"length"`
}

type MaValue struct {
	Ma5  float64 `json:"ma5"`
	Ma10 float64 `json:"ma10"`
	Ma20 float64 `json:"ma20"`
	Ma30 float64 `json:"ma30"`
	Ma60 float64 `json:"ma60"`
}

type BollingValue struct {
	LastBollingUp   float64         `json:"last_bolling_up"`
	LastBollingDown float64         `json:"last_bolling_down"`
	LastBollingMid  float64         `json:"last_bolling_mid"`
	LastPrice       float64         `json:"last_price"`
	ClosedPosition  BollingPosition `json:"closed_position"`
}

type KdjValue struct {
	LastKdjK float64 `json:"last_kdj_k"`
	LastKdjD float64 `json:"last_kdj_d"`
	LastKdjJ float64 `json:"last_kdj_j"`
}

type AnalyzeStockCodeResp struct {
	SuggestOperation StockSuggestOperation `json:"suggest_operation"`
	SuggestReason    string                `json:"suggest_reason,omitempty"`
	SuggestRange     string                `json:"suggest_range,omitempty"`
	SuggestPriority  int                   `json:"suggest_priority,omitempty"`
	MacdValue        *MacdValue            `json:"macd_value,omitempty"`
	MaValue          *MaValue              `json:"ma_value,omitempty"`
	BollingValue     *BollingValue         `json:"bolling_value,omitempty"`
	KdjValue         *KdjValue             `json:"kdj_value,omitempty"`
}

type SortMaOrderData []*MaOrderData

func (s SortMaOrderData) Len() int {
	return len(s)
}
func (s SortMaOrderData) Less(i, j int) bool {
	return s[i].MaPrice < s[j].MaPrice
}

func (s SortMaOrderData) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortMaOrderData) GetOrder() string {
	list := make([]string, len(s))
	for i, item := range s {
		list[i] = string(item.MaType)
	}
	return strings.Join(list, " < ")
}

func (f *MaFilter) Filter(item *FilterStockCodeItem) bool {
	if f == nil {
		return true
	}
	priceData := item.Result[StockStrategyMa]
	lastSortedPriceList := []*MaOrderData{
		{
			MaType:  StockMaType5,
			MaPrice: priceData.MaValue.Ma5,
		},
		{
			MaType:  StockMaType10,
			MaPrice: priceData.MaValue.Ma10,
		},
		{
			MaType:  StockMaType20,
			MaPrice: priceData.MaValue.Ma20,
		},
		{
			MaType:  StockMaType30,
			MaPrice: priceData.MaValue.Ma30,
		},
		{
			MaType:  StockMaType60,
			MaPrice: priceData.MaValue.Ma60,
		},
	}
	sort.Sort(SortMaOrderData(lastSortedPriceList))

	if f.Ma5Position > 0 && lastSortedPriceList[f.Ma5Position-1].MaType != StockMaType5 {
		return false
	}
	if f.Ma10Position > 0 && lastSortedPriceList[f.Ma10Position-1].MaType != StockMaType10 {
		return false
	}
	if f.Ma20Position > 0 && lastSortedPriceList[f.Ma20Position-1].MaType != StockMaType20 {
		return false
	}
	if f.Ma30Position > 0 && lastSortedPriceList[f.Ma30Position-1].MaType != StockMaType30 {
		return false
	}
	if f.Ma60Position > 0 && lastSortedPriceList[f.Ma60Position-1].MaType != StockMaType60 {
		return false
	}
	return true
}

func (f *MacdFilter) Filter(item *FilterStockCodeItem) bool {
	if f == nil {
		return true
	}
	priceData := item.Result[StockStrategyMacd]
	if f.MaxLastDif > 0.0 || f.MaxLastDea > 0.0 || f.MinLength > 0 {
		if priceData.MacdValue == nil {
			return false
		}
	}
	if f.MaxLastDif > 0.0 && priceData.MacdValue.LastDif > f.MaxLastDif {
		return false
	}
	if f.MaxLastDea > 0.0 && priceData.MacdValue.LastDea > f.MaxLastDea {
		return false
	}
	if f.MinLength > 0 && priceData.MacdValue.Length < f.MinLength {
		return false
	}
	return true
}

func (f *BollingFilter) Filter(item *FilterStockCodeItem) bool {
	if f == nil {
		return true
	}
	priceData := item.Result[StockStrategyBolling]
	if len(f.BollingPosition) > 0 && priceData.BollingValue.ClosedPosition != f.BollingPosition {
		return false
	}
	return true
}

func (f *KdjFilter) Filter(item *FilterStockCodeItem) bool {
	if f == nil {
		return true
	}
	priceData := item.Result[StockStrategyKdj]
	if f.MaxLastK > 0.0 && priceData.KdjValue.LastKdjK > f.MaxLastK {
		return false
	}
	if f.MaxLastD > 0.0 && priceData.KdjValue.LastKdjD > f.MaxLastD {
		return false
	}
	if f.MaxLastJ > 0.0 && priceData.KdjValue.LastKdjJ > f.MaxLastJ {
		return false
	}
	return true
}

type AnalyzeTrendCodeReq struct {
	Code      string    `json:"code"`
	StartDate string    `json:"start_date,omitempty"`
	EndDate   string    `json:"end_date,omitempty"`
	KLineType KLineType `json:"k_line_type"`
}

type AnalyzeTrendCodeResp struct {
	TrendFractal        []*FractalItem         `json:"trend_fractal"`
	PriceData           []*PriceItem           `json:"price_data"`
	PivotData           []*PivotItem           `json:"pivot_data"`
	DivergencePointData []*DivergencePointItem `json:"divergence_point_data"`
}

type PriceItem struct {
	Date       string  `json:"date"`
	PriceHigh  float64 `json:"price_high"`
	PriceLow   float64 `json:"price_low"`
	PriceOpen  float64 `json:"price_open"`
	PriceClose float64 `json:"price_close"`
	Amount     int64   `json:"amount" gorm:"column:amount"`
}

type FractalItem struct {
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	Class      ClassType `json:"class"`
	PriceStart float64   `json:"price_start"`
	PriceEnd   float64   `json:"price_end"`
}

type PivotItem struct {
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	PriceLow  float64 `json:"price_low"`
	PriceHigh float64 `json:"price_high"`
}

type DivergencePointItem struct {
	Date      string  `json:"date"`
	PointType string  `json:"point_type"`
	Price     float64 `json:"price"`
}
