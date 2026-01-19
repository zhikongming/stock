package model

type EMGetRemoteStockBasicResp struct {
	Version string              `json:"version"`
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Code    int                 `json:"code"`
	Result  *EMStockBasicResult `json:"result"`
}

type EMStockBasicResult struct {
	Data []*EMStockBasicData `json:"data"`
}

type EMStockBasicData struct {
	SecretaryNameAbbr string `json:"SECURITY_NAME_ABBR"`
	ListingDate       string `json:"LISTING_DATE"`
}

type EMStockRelationResp struct {
	Data map[string]interface{} `json:"data"`
}

type EMGetRemoteStockDailyResp struct {
	Data *EMStockDailyData `json:"data"`
}

type EMStockDailyData struct {
	Code      string   `json:"code"`
	Market    int      `json:"market"`
	Name      string   `json:"name"`
	Decimal   int      `json:"decimal"`
	Dktotal   int      `json:"dktotal"`
	PreKPrice float64  `json:"preKPrice"`
	Klines    []string `json:"klines"`
}

type EMGetRemoteStockIndustryResp struct {
	Data *EMStockIndustryItem `json:"data"`
}

type EMStockIndustryItem struct {
	Total int                               `json:"total"`
	Diff  map[string]map[string]interface{} `json:"diff"`
}

type IndustryItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type StockItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type WrapStockItem struct {
	IndustryCode string
	Err          error
	StockItem    []*StockItem
}

type WrapIndustryStockPriceItem struct {
	Err error
}

type EMGetRemoteFundFlowResp struct {
	Data *EMFundFlowData `json:"data"`
}

type EMFundFlowData struct {
	Total int                      `json:"total"`
	Diff  []map[string]interface{} `json:"diff"`
}

type FundFlowData struct {
	Code                     string  `json:"code"`
	Name                     string  `json:"name"`
	MainInflowAmount         int64   `json:"main_inflow_amount"`
	ExtremeLargeInflowAmount int64   `json:"extreme_large_inflow_amount"`
	LargeInflowAmount        int64   `json:"large_inflow_amount"`
	MediumInflowAmount       int64   `json:"medium_inflow_amount"`
	SmallInflowAmount        int64   `json:"small_inflow_amount"`
	PriceClose               float64 `json:"price_close"`
	Date                     string  `json:"date"`
}

type EMGetRemoteDailyFundFlowResp struct {
	Data *EMDailyFundFlowData `json:"data"`
}

type EMDailyFundFlowData struct {
	Code   string   `json:"code"`
	Name   string   `json:"name"`
	Market int      `json:"market"`
	Klines []string `json:"klines"`
}

type WrapFundFlowData struct {
	Err          error
	StockCode    string          `json:"stock_code"`
	FundFlowData []*FundFlowData `json:"fund_flow_data"`
}
