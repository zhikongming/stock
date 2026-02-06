package model

import "time"

const (
	CodeTypeStock    string = "stock"
	CodeTypeIndustry string = "industry"
)

type AddWatcherReq struct {
	Name          string   `json:"name"`
	StockCodeList []string `json:"stock_code_list"`
}

type GetWatchersReq struct {
	ID int64 `json:"id" query:"id"`
}

type GetWatchersResp struct {
	Data []*Watcher `json:"watchers"`
}

type DeleteWatcherReq struct {
	ID int64 `json:"id"`
}

type MultiCodeInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Watcher struct {
	ID         uint             `json:"id" gorm:"primaryKey"`
	Name       string           `json:"name" gorm:"column:name"`
	Stocks     []*MultiCodeInfo `json:"stocks" gorm:"column:stocks"`
	StockType  int              `json:"stock_type" gorm:"column:stock_type"`
	UpdateTime time.Time        `json:"update_time" gorm:"column:update_time"`
}

type MultiCodeInfoSorter []*MultiCodeInfo

func (s MultiCodeInfoSorter) Len() int {
	return len(s)
}
func (s MultiCodeInfoSorter) Less(i, j int) bool {
	// 先按类型排序，行业在股票之前
	if s[i].Type != s[j].Type {
		return s[i].Type == CodeTypeIndustry
	}
	return false
}
func (s MultiCodeInfoSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
