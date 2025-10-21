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
