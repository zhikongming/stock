package service

import (
	"context"
	"fmt"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func NewStrategyParser(ctx context.Context, strategy *model.AddSubscribeStrategyReq) (StrategyParser, error) {
	switch strategy.StrategyType {
	case model.StrategyTypeIndustryRateChange:
		return &IndustryRateChangeStrategyParser{
			ctx:          ctx,
			IndustryCode: strategy.IndustryCode,
			Days:         strategy.Days,
			RateChange:   strategy.RateChange,
		}, nil
	case model.StrategyTypeStockRateChange:
		return &StockRateChangeStrategyParser{
			ctx:        ctx,
			StockCode:  strategy.StockCode,
			Days:       strategy.Days,
			RateChange: strategy.RateChange,
		}, nil
	case model.StrategyTypeStockPriceChange:
		return &PriceChangeStrategyParser{
			ctx:             ctx,
			StockCode:       strategy.StockCode,
			PriceChange:     strategy.PriceChange,
			PriceChangeType: strategy.PriceChangeType,
		}, nil
	default:
		return nil, fmt.Errorf("unknown strategy type: %d", strategy.StrategyType)
	}
}

type IndustryRateChangeStrategyParser struct {
	StrategyParserBasic
	ctx          context.Context
	IndustryCode string
	Days         int
	RateChange   float64 `json:"rate_change"`
}

func (p *IndustryRateChangeStrategyParser) Parse() (*StrategyParseResult, error) {
	// 检查行业代码是否为空
	if p.IndustryCode == "" {
		return nil, fmt.Errorf("industry code is empty")
	}
	if p.Days <= 0 {
		return nil, fmt.Errorf("days must be greater than 0")
	}
	// 检查行业代码是否存在
	industryData, err := dal.GetStockIndustry(p.ctx, p.IndustryCode)
	if err != nil {
		return nil, fmt.Errorf("industry code %s not found: %w", p.IndustryCode, err)
	}
	// 获取时间区间内的行业价格数据
	req := &model.GetIndustryTrendDataReq{
		Days:         p.Days,
		SyncPrice:    false,
		IndustryCode: "",
	}
	industryPriceTrendList, err := GetIndustryTrendDetailByIndustryCode(p.ctx, req, p.IndustryCode)
	if err != nil {
		return nil, err
	}
	if len(industryPriceTrendList) != 1 {
		return nil, fmt.Errorf("industry price trend list is empty")
	}
	industryPriceTrend := industryPriceTrendList[0]
	// 计算行业价格变化率
	rateChange := (industryPriceTrend.PriceTrendList[len(industryPriceTrend.PriceTrendList)-1].Price - 1) * 100
	// 检查行业价格变化率是否符合策略, 波动可以为正也可以为负
	result := false
	if (p.RateChange > 0 && rateChange > p.RateChange) || (p.RateChange < 0 && rateChange < p.RateChange) {
		result = true
	}
	var strategyResult string
	if result {
		strategyResult = fmt.Sprintf("行业 %s(%s) 价格变化率 %s, 符合策略 %s",
			industryData.Name, industryData.Code, p.GetRateChangeMessage(rateChange), p.GetRateChangeMessage(p.RateChange))
	} else {
		strategyResult = fmt.Sprintf("行业 %s(%s) 价格变化率 %s, 不符合策略 %s",
			industryData.Name, industryData.Code, p.GetRateChangeMessage(rateChange), p.GetRateChangeMessage(p.RateChange))
	}
	return &StrategyParseResult{
		Result:         result,
		StrategyResult: strategyResult,
		Code:           fmt.Sprintf("%s(%s)", industryData.Name, industryData.Code),
		LastDate:       industryPriceTrend.PriceTrendList[len(industryPriceTrend.PriceTrendList)-1].DateString,
	}, nil
}

func (p *IndustryRateChangeStrategyParser) ToSubscribeStrategyDetail() string {
	return fmt.Sprintf("行业价格变化率%s", p.GetRateChangeMessage(p.RateChange))
}

type StockRateChangeStrategyParser struct {
	ctx        context.Context
	StockCode  string
	Days       int
	RateChange float64 `json:"rate_change"`
}

func (p *StockRateChangeStrategyParser) Parse() (*StrategyParseResult, error) {
	// 检查股票代码是否为空
	if p.StockCode == "" {
		return nil, fmt.Errorf("stock code is empty")
	}
	if p.Days <= 0 {
		return nil, fmt.Errorf("days must be greater than 0")
	}
	// 检查股票代码是否存在
	stockData, err := dal.GetStockCodeByCode(p.ctx, p.StockCode)
	if err != nil {
		return nil, fmt.Errorf("stock code %s not found: %w", p.StockCode, err)
	}
	// 获取时间区间内的股价数据
	limit := p.Days + 1
	stockPriceList, err := dal.GetLastNStockPrice(p.ctx, p.StockCode, "", limit)
	if err != nil {
		return nil, fmt.Errorf("get last %d stock price failed: %w", limit, err)
	}
	if len(stockPriceList) < limit {
		return nil, fmt.Errorf("not enough stock price data, expect %d, got %d", limit, len(stockPriceList))
	}
	// 计算股票价格变化率
	firstPrice := stockPriceList[len(stockPriceList)-1]
	lastPrice := stockPriceList[0]
	rateChange := (lastPrice.PriceClose - firstPrice.PriceClose) / firstPrice.PriceClose * 100
	// 检查股票价格变化率是否符合策略, 波动可以为正也可以为负
	result := false
	if (p.RateChange > 0 && rateChange > p.RateChange) || (p.RateChange < 0 && rateChange < p.RateChange) {
		result = true
	}
	var strategyResult string
	if result {
		strategyResult = fmt.Sprintf("股票 %s(%s) 价格变化率 %s, 符合策略 %s",
			stockData.CompanyName, stockData.CompanyCode, p.getRateChangeMessage(rateChange), p.getRateChangeMessage(p.RateChange))
	} else {
		strategyResult = fmt.Sprintf("股票 %s(%s) 价格变化率 %s, 不符合策略 %s",
			stockData.CompanyName, stockData.CompanyCode, p.getRateChangeMessage(rateChange), p.getRateChangeMessage(p.RateChange))
	}
	return &StrategyParseResult{
		Result:         result,
		StrategyResult: strategyResult,
		Code:           fmt.Sprintf("%s(%s)", stockData.CompanyName, stockData.CompanyCode),
		LastDate:       utils.FormatDate(lastPrice.Date),
	}, nil
}

func (p *StockRateChangeStrategyParser) getRateChangeMessage(priceChange float64) string {
	if priceChange >= 0 {
		return fmt.Sprintf("上涨%.2f%%", priceChange)
	}
	return fmt.Sprintf("下跌%.2f%%", -priceChange)
}

func (p *StockRateChangeStrategyParser) ToSubscribeStrategyDetail() string {
	return fmt.Sprintf("个股股价波动率%s", p.getRateChangeMessage(p.RateChange))
}

type PriceChangeStrategyParser struct {
	ctx             context.Context
	StockCode       string
	PriceChange     float64               `json:"price_change"`
	PriceChangeType model.PriceChangeType `json:"price_change_type"`
}

func (p *PriceChangeStrategyParser) Parse() (*StrategyParseResult, error) {
	if p.StockCode == "" {
		return nil, fmt.Errorf("stock code is empty")
	}
	stockData, err := dal.GetStockCodeByCode(p.ctx, p.StockCode)
	if err != nil {
		return nil, err
	}
	// 获取最新的股价信息
	price, err := dal.GetLastStockPrice(p.ctx, p.StockCode)
	if err != nil {
		return nil, fmt.Errorf("get last stock price failed: %w", err)
	}
	// 检查股价是否符合策略
	result := false
	if p.PriceChangeType == model.PriceChangeTypeGreater && price.PriceClose >= p.PriceChange {
		result = true
	} else if p.PriceChangeType == model.PriceChangeTypeLess && price.PriceClose <= p.PriceChange {
		result = true
	}
	// 构建策略结果字符串
	var strategyResult string
	if result {
		strategyResult = fmt.Sprintf("最新收盘价 %.2f 符合%s %.2f 的策略, 距离目标已超过 %.2f%%",
			price.PriceClose, p.PriceChangeType.String(), p.PriceChange, utils.Float64KeepDecimal((price.PriceClose-p.PriceChange)/p.PriceChange*100, 2))
	} else {
		strategyResult = fmt.Sprintf("最新收盘价 %.2f 不符合%s %.2f 的策略, 距离目标还差 %.2f%%",
			price.PriceClose, p.PriceChangeType.String(), p.PriceChange, utils.Float64KeepDecimal((p.PriceChange-price.PriceClose)/p.PriceChange*100, 2))
	}

	return &StrategyParseResult{
		Result:         result,
		StrategyResult: strategyResult,
		Code:           fmt.Sprintf("%s(%s)", stockData.CompanyName, stockData.CompanyCode),
		LastDate:       utils.FormatDate(price.Date),
	}, nil
}

func (p *PriceChangeStrategyParser) ToSubscribeStrategyDetail() string {
	return fmt.Sprintf("个股股价目标收盘价%s %.2f", p.PriceChangeType.String(), p.PriceChange)
}
