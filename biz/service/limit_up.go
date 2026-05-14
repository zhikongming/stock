package service

import (
	"context"
	"math"
	"sort"
	"strings"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	MaxLimitUpReportJobNum = 50
	GetLastNStockPriceNum  = 15
)

func GetLimitUpReport(ctx context.Context) ([][]*model.LimitUpReportItem, error) {
	// 获取所有的股票信息
	allStockList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}
	// 获取所有的板块信息
	allIndustryList, err := dal.GetAllStockIndustry(ctx)
	if err != nil {
		return nil, err
	}
	industryMap := make(map[string]string)
	for _, industry := range allIndustryList {
		industryMap[industry.Code] = industry.Name
	}
	allIndustryRelationList, err := dal.GetAllStockIndustryRelation(ctx)
	if err != nil {
		return nil, err
	}
	stockMap := make(map[string]string)
	for _, industryRelation := range allIndustryRelationList {
		stockMap[industryRelation.CompanyCode] = industryMap[industryRelation.IndustryCode]
	}

	// 需要使用并发来计算, 以减少耗时
	jobList := make([]func() (interface{}, error), 0)
	for _, stockCode := range allStockList {
		jobList = append(jobList, func(stockCode *dal.StockCode) func() (interface{}, error) {
			return func() (interface{}, error) {
				stockPriceList, err := dal.GetLastNStockPrice(ctx, stockCode.CompanyCode, "", GetLastNStockPriceNum)
				if err != nil {
					return nil, err
				}
				if len(stockPriceList) != GetLastNStockPriceNum {
					return nil, nil
				}
				// 倒序开始处理
				count := CalculateLimitUpCount(stockPriceList)
				data := &model.LimitUpReportItem{
					Code:         stockCode.CompanyCode,
					Name:         stockCode.CompanyName,
					Count:        count,
					IndustryName: stockMap[stockCode.CompanyCode],
				}
				return data, nil
			}
		}(stockCode))
	}
	// 执行并发任务
	dataList, err := utils.ConcurrentActuator(jobList, MaxVolumeReportJobNum)
	if err != nil {
		return nil, err
	}
	maxCount := 0
	var ret []*model.LimitUpReportItem
	for _, item := range dataList {
		if item != nil {
			d := item.(*model.LimitUpReportItem)
			if d.Count > 0 {
				ret = append(ret, d)
				if d.Count > maxCount {
					maxCount = d.Count
				}
			}
		}
	}
	// 构建返回数据
	reportList := make([][]*model.LimitUpReportItem, 0, maxCount+1)
	for i := 0; i <= maxCount; i++ {
		reportList = append(reportList, []*model.LimitUpReportItem{})
	}
	for _, item := range ret {
		count := item.Count
		reportList[count] = append(reportList[count], item)
	}
	// 对每个连板数量的数据排序
	for _, group := range reportList {
		sort.Sort(model.LimitUpReportItemSorter(group))
	}
	return reportList, nil
}

func CalculateLimitUpCount(stockPriceList []*dal.StockPrice) int {
	if len(stockPriceList) < 2 {
		return 0
	}
	rate := GetLimitUpRate(utils.GetStockCodeNumber(stockPriceList[0].CompanyCode))
	count := 0
	for idx := 0; idx < len(stockPriceList)-1; idx++ {
		current := stockPriceList[idx]
		previous := stockPriceList[idx+1]
		if IsLimitUpWithRate(previous.PriceClose, current.PriceClose, rate) {
			count++
		} else {
			break
		}
	}
	return count
}

func IsLimitUpWithRate(prevClose, current float64, rate float64) bool {
	prevCloseInCents := int(math.Round(prevClose * 100))
	currentInCents := int(math.Round(current * 100))
	limitUpInCents := int(math.Round(float64(prevCloseInCents) * rate))
	return currentInCents >= limitUpInCents
}

func GetLimitUpRate(stockCode string) float64 {
	// 创业板（300/301开头）
	if strings.HasPrefix(stockCode, "30") || strings.HasPrefix(stockCode, "301") {
		return 1.20
	}
	// 科创板（688开头）
	if strings.HasPrefix(stockCode, "688") {
		return 1.20
	}
	// 北交所（8/43开头）
	if strings.HasPrefix(stockCode, "8") || strings.HasPrefix(stockCode, "43") {
		return 1.30
	}
	// 主板（其余代码）
	return 1.10
}
