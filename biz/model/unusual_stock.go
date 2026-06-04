package model

type UnusualType int

const (
	UnusualTypeNormal UnusualType = iota
	UnusualTypeSpecial
	UnusualTypeMarketRisk
)

type UnusualStock struct {
	Code          string      `json:"code"`           // 股票代码
	Name          string      `json:"name"`           // 股票名称
	Type          UnusualType `json:"type"`           // 异常类型: 0: 异常波动, 1: 严重异常波动, 2: 交易所风险提示
	StartDate     string      `json:"start_date"`     // 异常开始日期
	EndDate       string      `json:"end_date"`       // 异常结束日期
	NoticeDate    string      `json:"notice_date"`    // 通知日期
	UnusualType   string      `json:"unusual_type"`   // 异常类型
	UnusualReason string      `json:"unusual_reason"` // 异常原因
}
