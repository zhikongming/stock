package model

type GetSimilarCompanyReq struct {
	CompanyName string `json:"company_name"`
}

type GetSimilarCompanyResp struct {
	SimilarCompanies []*SimilarCompany `json:"similar_companies"`
}
type SimilarCompany struct {
	CompanyName         string  `json:"company_name"`
	SimilarityScore     float64 `json:"similarity_score"`
	SimilarityReason    string  `json:"similarity_reason"`
	BusinessDescription string  `json:"business_description"`
}

type GetVolumePriceReq struct {
	CompanyName   string       `json:"company_name"`
	StockDataList []*StockData `json:"stock_data_list"`
}

type StockData struct {
	Date       string `json:"date"`
	OpenPrice  string `json:"open_price"`
	ClosePrice string `json:"close_price"`
	HighPrice  string `json:"high_price"`
	LowPrice   string `json:"low_price"`
	Volume     string `json:"volume"`
}

type GetVolumePriceResp struct {
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	IsSafe         string `json:"is_safe"`
	AnalysisResult string `json:"analysis_result"`
}
