package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func AnalyzeThirdBuyCode(ctx context.Context, req *model.AnalyzeThirdBuyCodeReq) (*model.ThirdBuyCodePeriod, error) {
	if req.StockCode == "" {
		return nil, fmt.Errorf("bad request")
	}
	// 获取股票的数据信息
	var stockPriceList []*dal.StockPrice
	var err error
	if req.StartDate != "" {
		// 根据起始日期获取股价的数据信息
		stockPriceList, err = dal.GetStockPriceByDate(ctx, req.StockCode, req.StartDate, "", 0)
		if err != nil {
			return nil, err
		}
	} else if req.Days > 0 {
		// 根据天数获取股价的数据信息
		stockPriceList, err = dal.GetLastNStockPrice(ctx, req.StockCode, "", req.Days)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("bad request")
	}
	if len(stockPriceList) == 0 {
		return nil, fmt.Errorf("no stock price data")
	}
	// 格式化数据, 并分析第三类买点的三元数组
	// 将倒序变为正序
	stockPriceList = utils.ListSwap(stockPriceList)
	// 格式化数据
	codePriceList := formatStockPriceList(stockPriceList)
	// 分析第三类买点的三元数组
	thirdBuyCodePeriod := AnalyzeThirdBuyPeriod(codePriceList)
	if thirdBuyCodePeriod == nil {
		return nil, fmt.Errorf("no third buy code period")
	}
	return thirdBuyCodePeriod, nil
}

func formatStockPriceList(stockPriceList []*dal.StockPrice) []*model.CodePrice {
	var codePriceList []*model.CodePrice
	for _, stockPrice := range stockPriceList {
		codePriceList = append(codePriceList, &model.CodePrice{
			Code:  stockPrice.CompanyCode,
			Date:  utils.FormatDate(stockPrice.Date),
			Price: stockPrice.PriceClose,
		})
	}
	return codePriceList
}

func AnalyzeThirdBuyPeriod(codePriceList []*model.CodePrice) *model.ThirdBuyCodePeriod {
	ret := &model.ThirdBuyCodePeriod{}
	length := len(codePriceList)
	// 分析上涨阶段
	upStartIdx, upEndIdx := getUpPeriod(codePriceList)
	if upStartIdx == 0 && upEndIdx == 0 {
		return nil
	}
	ret.UpPeriod = &model.ThirdBuyPeriod{
		StartDate:  codePriceList[upStartIdx].Date,
		StartPrice: codePriceList[upStartIdx].Price,
		EndDate:    codePriceList[upEndIdx].Date,
		EndPrice:   codePriceList[upEndIdx].Price,
		Rate:       utils.Float64KeepDecimal((codePriceList[upEndIdx].Price-codePriceList[upStartIdx].Price)/codePriceList[upStartIdx].Price*100, 2),
	}
	// 分析回调阶段
	downStartIdx, downEndIdx := getDownPeriod(codePriceList, upEndIdx)
	ret.DownPeriod = &model.ThirdBuyPeriod{
		StartDate:  codePriceList[downStartIdx].Date,
		StartPrice: codePriceList[downStartIdx].Price,
		EndDate:    codePriceList[downEndIdx].Date,
		EndPrice:   codePriceList[downEndIdx].Price,
		Rate:       utils.Float64KeepDecimal((codePriceList[downStartIdx].Price-codePriceList[downEndIdx].Price)/codePriceList[downStartIdx].Price*100, 2),
	}
	// 分析重新上涨阶段
	reupEndIdx := length - 1
	ret.ReupPeriod = &model.ThirdBuyPeriod{
		StartDate:  codePriceList[downEndIdx].Date,
		StartPrice: codePriceList[downEndIdx].Price,
		EndDate:    codePriceList[reupEndIdx].Date,
		EndPrice:   codePriceList[reupEndIdx].Price,
		Rate:       utils.Float64KeepDecimal((codePriceList[reupEndIdx].Price-codePriceList[downEndIdx].Price)/codePriceList[downEndIdx].Price*100, 2),
	}
	// 分析当前操作的利润空间
	ret.FinalPeriod = &model.FinalThirdBuyPeriod{
		StartDate:  codePriceList[length-1].Date,
		StartPrice: codePriceList[length-1].Price,
		EndPrice:   codePriceList[upEndIdx].Price,
		Rate:       utils.Float64KeepDecimal((codePriceList[upEndIdx].Price-codePriceList[length-1].Price)/codePriceList[length-1].Price*100, 2),
	}
	return ret
}

func getUpPeriod(codePriceList []*model.CodePrice) (int, int) {
	n := len(codePriceList)
	if n < 2 {
		return 0, 0
	}

	minPriceIdx := 0 // 历史最低价下标
	minPrice := codePriceList[0].Price
	maxProfit := 0.0
	buyIdx, sellIdx := 0, 0 // 记录最佳买卖点

	for i := 1; i < n; i++ {
		if codePriceList[i].Price < minPrice {
			minPrice = codePriceList[i].Price
			minPriceIdx = i
		} else {
			profit := codePriceList[i].Price - minPrice
			if profit > maxProfit {
				maxProfit = profit
				buyIdx = minPriceIdx
				sellIdx = i
			}
		}
	}
	return buyIdx, sellIdx
}

func getDownPeriod(codePriceList []*model.CodePrice, startIdx int) (int, int) {
	n := len(codePriceList) - 1
	if startIdx >= n {
		return startIdx, startIdx
	}
	minPrice := codePriceList[startIdx].Price
	minPriceIdx := startIdx
	for i := startIdx + 1; i <= n; i++ {
		if codePriceList[i].Price < minPrice {
			minPrice = codePriceList[i].Price
			minPriceIdx = i
		}
	}
	return startIdx, minPriceIdx
}

func FilterThirdBuyCode(ctx context.Context, req *model.FilterThirdBuyCodeReq) (*model.FilterThirdBuyCodePeriodResp, error) {
	var industryRelationList []*dal.StockIndustryRelation
	var err error
	industryMap := make(map[string]string)
	if req.IndustryName == "" {
		industryRelationList, err = dal.GetAllStockIndustryRelation(ctx)
		if err != nil {
			return nil, err
		}
		industryList, err := dal.GetAllStockIndustry(ctx)
		if err != nil {
			return nil, err
		}
		for _, industry := range industryList {
			industryMap[industry.Code] = industry.Name
		}
	} else {
		// 根据行业名称获取所有股票代码
		industryData, err := dal.GetStockIndustryByName(ctx, req.IndustryName)
		if err != nil {
			return nil, err
		}
		if industryData == nil {
			return nil, fmt.Errorf("industry not found")
		}
		industryMap[industryData.Code] = industryData.Name
		industryRelationList, err = dal.GetStockIndustryRelation(ctx, industryData.Code)
		if err != nil {
			return nil, err
		}
	}

	// 过滤符合条件的股票代码
	ret := []*model.FilterThirdBuyCodePeriodResult{}
	matchStockCodeList := []string{}
	for _, industryRelation := range industryRelationList {
		stockCode := industryRelation.CompanyCode
		period, err := AnalyzeThirdBuyCode(ctx, &model.AnalyzeThirdBuyCodeReq{
			StockCode: stockCode,
			StartDate: req.StartDate,
			Days:      req.Days,
		})
		if err != nil {
			continue
		}
		if !period.ValidFilter(req.ThresholdUp, req.ThresholdPullback, req.ThresholdDeviation, req.ThresholdProfit) {
			continue
		}
		matchStockCodeList = append(matchStockCodeList, stockCode)
		ret = append(ret, &model.FilterThirdBuyCodePeriodResult{
			ThirdBuyCodePeriod: *period,
			Code:               stockCode,
			IndustryName:       industryMap[industryRelation.IndustryCode],
		})
	}
	// 填充股票名称信息
	stockCodeList, err := dal.GetStockCodeByCodeList(ctx, matchStockCodeList)
	if err != nil {
		return nil, err
	}
	stockCodeMap := map[string]string{}
	for _, stockCode := range stockCodeList {
		stockCodeMap[stockCode.CompanyCode] = stockCode.CompanyName
	}
	for _, item := range ret {
		stockCode := stockCodeMap[item.Code]
		if stockCode == "" {
			continue
		}
		item.Name = stockCode
	}
	// 排序
	sort.Sort(model.SorterFilterThirdBuyCodePeriodResult(ret))
	return &model.FilterThirdBuyCodePeriodResp{
		Total: len(industryRelationList),
		Data:  ret,
	}, nil
}
