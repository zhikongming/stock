package model

type LocalStockDailyData struct {
	Date          string  `json:"date"`
	PriceHigh     float64 `json:"price_high"`
	PriceLow      float64 `json:"price_low"`
	PriceOpen     float64 `json:"price_open"`
	PriceClose    float64 `json:"price_close"`
	ChangePercent float64 `json:"change_percent"`
	Amount        int64   `json:"amount"`
	BollingUp     float64 `json:"bolling_up"`
	BollingDown   float64 `json:"bolling_down"`
	BollingMid    float64 `json:"bolling_mid"`
	Ma5           float64 `json:"ma5"`
	Ma10          float64 `json:"ma10"`
	Ma20          float64 `json:"ma20"`
	Ma30          float64 `json:"ma30"`
	Ma60          float64 `json:"ma60"`
}
