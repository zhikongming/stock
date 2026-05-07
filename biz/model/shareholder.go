package model

// SyncShareholderReq 同步股东数据请求
type SyncShareholderReq struct {
	Code string `json:"code"`
}

// GetShareholderReportReq 获取股东报告请求
type GetShareholderReportReq struct {
	Data []*ShareholderReportReq `json:"data"`
}

type ShareholderReportReq struct {
	Object    string `json:"object"`
	Operation string `json:"operation"`
}

// GetShareholderReportResp 获取股东报告响应
type ShareholderReportItem struct {
	Code   string                     `json:"code"`
	Name   string                     `json:"name"`
	Report *ShareholderAnalysisReport `json:"report"`
}

type ShareholderAnalysisReport struct {
	ReportDate            string                 `json:"report_date"`
	ShareholderNumber     int                    `json:"shareholder_number"`
	ShareholderNumberDiff int                    `json:"shareholder_number_diff"`
	TopShareholderList    []*ShareholderWithDiff `json:"top_shareholder_list"`
	AddShareholderList    []*ShareholderDiff     `json:"add_shareholder_list"`
	DelShareholderList    []*ShareholderDiff     `json:"del_shareholder_list"`
	ChangeShareholderList []*ShareholderDiff     `json:"change_shareholder_list"`
}

type ShareholderWithDiff struct {
	Shareholder
	Diff string `json:"diff"`
}

type ShareholderDiff struct {
	ShareholderName string `json:"shareholder_name"`
	Diff            string `json:"diff"`
}
