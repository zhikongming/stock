package service

import "fmt"

type StrategyParseResult struct {
	Result         bool   `json:"result"`
	StrategyResult string `json:"strategy_result"`
	Code           string `json:"code"`
	LastDate       string `json:"last_date"`
}

type StrategyParser interface {
	Parse() (*StrategyParseResult, error)
	ToSubscribeStrategyDetail() string
}

type StrategyParserBasic struct {
}

func (s *StrategyParserBasic) GetRateChangeMessage(priceChange float64) string {
	if priceChange >= 0 {
		return fmt.Sprintf("上涨%.2f%%", priceChange)
	}
	return fmt.Sprintf("下跌%.2f%%", -priceChange)
}
