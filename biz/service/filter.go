package service

import (
	"context"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func FilterStockCode(ctx context.Context, req model.FilterStockCodeReq) ([]*model.FilterStockCodeItem, error) {
	// 检查是否存在股票基础数据, 如果不存在就同步数据
	stockCodeList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}

	resultList := make([]*model.FilterStockCodeItem, 0)
	for _, stockCode := range stockCodeList {
		lastStockPrice, err := dal.GetLastStockPrice(ctx, stockCode.CompanyCode)
		if err != nil {
			return nil, err
		}

		result := make(map[model.StockStrategy]*model.AnalyzeStockCodeResp)
		req := model.AnalyzeStockCodeReq{
			Code: stockCode.CompanyCode,
			Date: req.Date,
		}
		result[model.StockStrategyMa], err = AnalyzeMa(ctx, req)
		if err != nil {
			return nil, err
		}
		result[model.StockStrategyBolling], err = AnalyzeBolling(ctx, req)
		if err != nil {
			return nil, err
		}
		result[model.StockStrategyMacd], err = AnalyzeMacd(ctx, req)
		if err != nil {
			return nil, err
		}
		result[model.StockStrategyKdj], err = AnalyzeKdj(ctx, req)
		if err != nil {
			return nil, err
		}
		resultList = append(resultList, &model.FilterStockCodeItem{
			Code:        stockCode.CompanyCode,
			CompanyName: stockCode.CompanyName,
			Result:      result,
			LastDate:    utils.FormatDate(lastStockPrice.Date),
		})
	}

	// 过滤数据
	filterResultList := make([]*model.FilterStockCodeItem, 0)
	for _, item := range resultList {
		if req.MaFilter.Filter(item) &&
			req.BollingFilter.Filter(item) &&
			req.MacdFilter.Filter(item) &&
			req.KdjFilter.Filter(item) {
			filterResultList = append(filterResultList, item)
		}
	}

	return filterResultList, nil
}
