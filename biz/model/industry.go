package model

import (
	"time"

	"github.com/zhikongming/stock/utils"
)

type GetIndustryBasicDataReq struct {
	IndustryCode string `json:"industry_code" query:"industry_code"`
}

type CodeBasic struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type IndustryBasicData struct {
	IndustryCode    string       `json:"industry_code"`
	IndustryName    string       `json:"industry_name"`
	CompanyCodeList []*CodeBasic `json:"company_code_list"`
}

type GetIndustryRelationDataReq struct {
	Days            int    `json:"days" query:"days"`
	IndustryCode    string `json:"industry_code" query:"industry_code"`
	IsSplitIndustry bool   `json:"is_split_industry" query:"is_split_industry"`
}

type GetIndustryTrendDataReq struct {
	Days         int    `json:"days" query:"days"`
	SyncPrice    bool   `json:"sync_price" query:"sync_price"`
	IndustryCode string `json:"industry_code" query:"industry_code"`
	EndDate      string `json:"end_date" query:"end_date"`
}

type GetIndustryTrendDataResp struct {
	IndustryPriceTrend []*IndustryPriceTrend `json:"industry_price_trend"`
	IndustryCodeTrend  []*IndustryCodeTrend  `json:"industry_code_trend"`
}

type IndustryPriceTrend struct {
	IndustryCode   string        `json:"industry_code"`
	IndustryName   string        `json:"industry_name"`
	PriceTrendList []*PriceTrend `json:"price_trend_list"`
}

type IndustryCodeTrend struct {
	StockCode      string        `json:"stock_code"`
	StockName      string        `json:"stock_name"`
	PriceTrendList []*PriceTrend `json:"price_trend_list"`
}

type PriceTrend struct {
	DateString string    `json:"date"`
	Diff       float64   `json:"diff"`
	Price      float64   `json:"price"`
	Date       time.Time `json:"-"`
}

type DatePrice struct {
	Date  string  `json:"date"`
	Price float64 `json:"price"`
}

type GetIndustryRelationDataResp struct {
	StartDate                 string                `json:"start_date"`
	EndDate                   string                `json:"end_date"`
	IndustryRelationList      []*IndustryRelation   `json:"industry_relation_list"`
	SplitIndustryRelationList [][]*IndustryRelation `json:"split_industry_relation_list"`
}

type IndustryRelation struct {
	IndustryCode      string  `json:"industry_code"`
	IndustryName      string  `json:"industry_name"`
	Correlation       float64 `json:"correlation"`
	CorrelationString string  `json:"correlation_string"`
}

// 认定为按照日期的顺序结构
func TransferDatePriceToPriceTrend(dList []*DatePrice) []*PriceTrend {
	data := make([]*PriceTrend, 0)
	length := len(dList)
	for i := 1; i < length; i++ {
		data = append(data, &PriceTrend{
			DateString: dList[i].Date,
			Price:      dList[i].Price,
			Diff:       utils.Float64KeepDecimal(100*(dList[i].Price-dList[i-1].Price)/dList[i-1].Price, 4),
		})
	}
	return data
}

type SortPriceTrend []*PriceTrend

func (s SortPriceTrend) Len() int {
	return len(s)
}
func (s SortPriceTrend) Less(i, j int) bool {
	return s[i].Date.Before(s[j].Date)
}

func (s SortPriceTrend) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type SortIndustryPriceTrend []*IndustryPriceTrend

func (s SortIndustryPriceTrend) Len() int {
	return len(s)
}
func (s SortIndustryPriceTrend) Less(i, j int) bool {
	if len(s[i].PriceTrendList) == 0 || len(s[j].PriceTrendList) == 0 {
		return false
	}
	return s[i].PriceTrendList[len(s[i].PriceTrendList)-1].Price > s[j].PriceTrendList[len(s[j].PriceTrendList)-1].Price
}

func (s SortIndustryPriceTrend) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type SortIndustryCodeTrend []*IndustryCodeTrend

func (s SortIndustryCodeTrend) Len() int {
	return len(s)
}
func (s SortIndustryCodeTrend) Less(i, j int) bool {
	if len(s[i].PriceTrendList) == 0 || len(s[j].PriceTrendList) == 0 {
		return false
	}
	return s[i].PriceTrendList[len(s[i].PriceTrendList)-1].Price > s[j].PriceTrendList[len(s[j].PriceTrendList)-1].Price
}

func (s SortIndustryCodeTrend) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type SortIndustryRelation []*IndustryRelation

func (s SortIndustryRelation) Len() int {
	return len(s)
}
func (s SortIndustryRelation) Less(i, j int) bool {
	return s[i].Correlation > s[j].Correlation
}

func (s SortIndustryRelation) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type CodeDiffPrice struct {
	Date  string
	Diff  float64
	Price float64
}
