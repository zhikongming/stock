package model

type LimitUpReportItem struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	Count        int    `json:"count"`
	IndustryName string `json:"industry_name"`
}
