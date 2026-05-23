package model

// 概念股票信息
type ConceptStockInfo struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Percent    float64 `json:"percent"`
	MaxPercent int     `json:"max_percent"`
}

// 概念响应结构
type ConceptResp struct {
	ID      uint                `json:"id"`
	Name    string              `json:"name"`
	Percent float64             `json:"percent"`
	Stocks  []*ConceptStockInfo `json:"stocks"`
}

// 获取概念列表请求
type GetConceptsReq struct {
	WithChange bool `query:"with_change"`
}

// 添加概念请求
type AddConceptReq struct {
	Name string `json:"name"`
}

// 删除概念请求
type DeleteConceptReq struct {
	ID int64 `json:"id"`
}

// 获取概念股票列表请求
type GetConceptStocksReq struct {
	ConceptID int64 `query:"concept_id"`
}

// 添加概念股票请求
type AddConceptStockReq struct {
	ConceptID int64  `json:"concept_id"`
	StockCode string `json:"stock_code"`
}

// 删除概念股票请求
type DeleteConceptStockReq struct {
	ConceptID int64  `json:"concept_id"`
	StockCode string `json:"stock_code"`
}

type WrapMinutePrice struct {
	Code string `json:"code"`
	*StockMinuteData
}
