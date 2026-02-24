package model

type BDGetRemoteStockDailyResp struct {
	Result *BDStockDailyResult `json:"Result"`
}

type BDStockDailyResult struct {
	NewMarketData *BDStockDailyData `json:"newMarketData"`
}

type BDStockDailyData struct {
	Headers    []string `json:"headers"`
	Keys       []string `json:"keys"`
	MarketData string   `json:"marketData"`
}
