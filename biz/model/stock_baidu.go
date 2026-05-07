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

type BDGetRemoteShareholderResp struct {
	Result *BDShareholdersResult `json:"Result"`
}

type BDShareholdersResult struct {
	BDShareholders *BDShareholders `json:"shareholders"`
	BDHoldShare    *BDHoldShare    `json:"holdShareInfo"`
}

type BDShareholders struct {
	BDShareholderList []*BDShareholder `json:"list"`
}

type BDShareholder struct {
	NumOrigin  string `json:"numOrigin"`
	Num        string `json:"num"`
	Change     string `json:"change"`
	Status     string `json:"status"`
	Price      string `json:"price"`
	ReportDate string `json:"reportDate"`
}

type BDHoldShare struct {
	Content *BDHoldShareContent `json:"content"`
}

type BDHoldShareContent struct {
	Tip  string             `json:"tip"`
	Body []*BDHoldShareItem `json:"body"`
}

type BDHoldShareItem struct {
	Holder     string `json:"holder"`
	HoldNum    string `json:"holdNum"`
	HoldPer    string `json:"holdPer"`
	HoldChange string `json:"holdChange"`
	Status     string `json:"status"`
}

type Top10Shareholder struct {
	Num             int            `json:"num"`
	ShareholderList []*Shareholder `json:"shareholder_list"`
}

type Shareholder struct {
	ShareholderName   string `json:"shareholder_name"`
	ShareholderNumber string `json:"shareholder_number"`
	ShareholderPer    string `json:"shareholder_per"`
}

type GetRemoteCompanyInfoResp struct {
	Result *BDCompanyInfoResult `json:"Result"`
}

type BDCompanyInfoResult struct {
	Code      string                `json:"code"`
	StockName string                `json:"stockName"`
	Content   *BDCompanyInfoContent `json:"content"`
}

type BDCompanyInfoContent struct {
	NewCompany *BDNewCompany `json:"newCompany"`
}

type BDNewCompany struct {
	BasicInfo *BDBasicInfo `json:"basicInfo"`
}

type BDBasicInfo struct {
	CompanyCode string `json:"companyCode"`
}
