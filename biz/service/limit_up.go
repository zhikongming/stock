package service

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

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
					LastDate:     stockPriceList[0].Date,
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
	// 过滤掉日期不正确的股票
	var lastDate time.Time
	for _, item := range ret {
		if item.LastDate.After(lastDate) {
			lastDate = item.LastDate
		}
	}
	var retNew []*model.LimitUpReportItem
	for _, item := range ret {
		if !item.LastDate.Before(lastDate) {
			retNew = append(retNew, item)
		}
	}
	// 构建返回数据
	reportList := make([][]*model.LimitUpReportItem, 0, maxCount+1)
	for i := 0; i <= maxCount; i++ {
		reportList = append(reportList, []*model.LimitUpReportItem{})
	}
	for _, item := range retNew {
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
	if stockPriceList[0].PriceClose <= 1.0 {
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
	// 1. 使用精度偏移量处理 (Precision Offset)
	// 金融计算中，为了处理 18.837 这种刚好在边缘的情况，
	// 我们可以计算出理论涨幅后，取 2 位小数的截断值。

	// 计算公式：floor(prevClose * (1 + rate) * 100 + 0.00001) / 100
	// 加上 0.00001 是为了防止浮点数表示 18.837 变成 18.836999999999 的误差
	limitUpPrice := math.Floor(prevClose*rate*100+0.00001) / 100

	// 2. 只要当前价大于或等于这个截断计算出的价格，即为涨停
	// 在 14.49 * 1.3 = 18.837 的情况下，limitUpPrice 会得到 18.83
	return current >= limitUpPrice
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
	// 北交所（已知前缀列表，根据实际情况持续补充）
	if strings.HasPrefix(stockCode, "8") || strings.HasPrefix(stockCode, "43") ||
		strings.HasPrefix(stockCode, "82") || strings.HasPrefix(stockCode, "83") ||
		strings.HasPrefix(stockCode, "87") || strings.HasPrefix(stockCode, "88") ||
		strings.HasPrefix(stockCode, "920") || strings.HasPrefix(stockCode, "921") ||
		strings.HasPrefix(stockCode, "922") {
		return 1.30
	}
	// 主板（其余代码）
	return 1.10
}
