package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

// SyncShareholder 同步股东数据
func SyncShareholder(ctx context.Context, req *model.SyncShareholderReq) error {
	stockList := make([]string, 0)
	// 如果股票代码为空，返回则考虑是否需要同步所有股票的股东数据
	if req.Code != "" {
		stockCode, err := dal.GetStockCodeByCode(ctx, req.Code)
		if err != nil {
			return err
		}
		if stockCode == nil {
			return fmt.Errorf("stock code not found for code: %s", req.Code)
		}
		stockList = append(stockList, stockCode.CompanyCode)
	} else {
		allStockCodeList, err := dal.GetAllStockCode(ctx)
		if err != nil {
			return err
		}
		for _, stockCode := range allStockCodeList {
			stockList = append(stockList, stockCode.CompanyCode)
		}
	}

	// 获取时间, 并检查最新的报告是否存在
	reportDate := utils.GetShareholderReportDate(time.Now())
	preReportDate := utils.GetPreShareholderReportDate(reportDate)
	cozeCache := GetCozeCache()
	for _, stockCode := range stockList {
		// 获取这两份报告, 如果没有则拉取并缓存
		curCache, err := dal.GetCacheByTypeDate(ctx, stockCode, dal.CacheTypeShareholderReport, reportDate)
		if err != nil {
			return err
		}
		if curCache == nil {
			// 拉取并缓存最新的报告
			curCache, err = cozeCache.GetAndSetShareholderReport(ctx, stockCode, reportDate)
			if err != nil {
				return err
			}
		}

		preCache, err := dal.GetCacheByTypeDate(ctx, stockCode, dal.CacheTypeShareholderReport, preReportDate)
		if err != nil {
			return err
		}
		if preCache == nil {
			// 拉取并缓存上一期报告
			preCache, err = cozeCache.GetAndSetShareholderReport(ctx, stockCode, preReportDate)
			if err != nil {
				return err
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// GetShareholderReport 获取股东报告
func GetShareholderReport(ctx context.Context, req *model.GetShareholderReportReq) ([]*model.ShareholderReportItem, error) {
	// 初始化过滤器
	filterList := make([]ShareholderFilter, 0)
	for _, r := range req.Data {
		filter := GetShareholderFilter(r)
		if filter == nil {
			return nil, fmt.Errorf("invalid filter type: %s", r.Operation)
		}
		filterList = append(filterList, filter)
	}
	// 过滤数据
	stockList := make([]string, 0)
	stockMap := make(map[string]string)
	allStockCodeList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}
	for _, stockCode := range allStockCodeList {
		stockList = append(stockList, stockCode.CompanyCode)
		stockMap[stockCode.CompanyCode] = stockCode.CompanyName
	}

	// 获取时间, 并获取最新的报告
	reportDate := utils.GetShareholderReportDate(time.Now())
	allReportList, err := dal.GetAllCacheByTypeDate(ctx, dal.CacheTypeShareholderReport, reportDate)
	if err != nil {
		return nil, err
	}
	allReportMap := make(map[string]*dal.Cache)
	for _, cache := range allReportList {
		allReportMap[cache.DataKey] = cache
	}
	preReportDate := utils.GetPreShareholderReportDate(reportDate)
	allPreReportList, err := dal.GetAllCacheByTypeDate(ctx, dal.CacheTypeShareholderReport, preReportDate)
	if err != nil {
		return nil, err
	}
	allPreReportMap := make(map[string]*dal.Cache)
	for _, cache := range allPreReportList {
		allPreReportMap[cache.DataKey] = cache
	}

	matchMap := make(map[string]*model.ShareholderAnalysisReport)
	for _, stockCode := range stockList {
		// 获取这两份报告, 如果没有则拉取并缓存
		curCache, ok := allReportMap[stockCode]
		if !ok {
			continue
		}

		preCache, ok := allPreReportMap[stockCode]
		if !ok {
			continue
		}

		var curReport *model.Top10Shareholder
		var preReport *model.Top10Shareholder
		err = json.Unmarshal([]byte(curCache.DataValue), &curReport)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(preCache.DataValue), &preReport)
		if err != nil {
			return nil, err
		}
		shareholderAnalysisInfo := ToShareholderAnalysisReport(curReport, preReport)
		match := true
		for _, filter := range filterList {
			if !filter.Filter(ctx, shareholderAnalysisInfo) {
				match = false
				break
			}
		}
		if match {
			matchMap[stockCode] = shareholderAnalysisInfo
		}
	}
	itemList := make([]*model.ShareholderReportItem, 0)
	for stockCode, report := range matchMap {
		itemList = append(itemList, &model.ShareholderReportItem{
			Code:   stockCode,
			Name:   stockMap[stockCode],
			Report: report,
		})
	}
	return itemList, nil
}
