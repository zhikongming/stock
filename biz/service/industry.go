package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

type WrapJob struct {
	Data      *dal.StockPrice
	StockCode string
}

type WrapJobs struct {
	Data      []*dal.StockPrice
	StockCode string
}

func GetIndustryBasicData(ctx context.Context, req *model.GetIndustryBasicDataReq) ([]*model.IndustryBasicData, error) {
	industryList := make([]*dal.StockIndustry, 0)
	var err error
	// 获取板块数据
	if len(req.IndustryCode) > 0 {
		industry, err := dal.GetStockIndustry(ctx, req.IndustryCode)
		if err != nil {
			return nil, err
		}
		if industry != nil {
			industryList = append(industryList, industry)
		}
	} else {
		industryList, err = dal.GetAllStockIndustry(ctx)
		if err != nil {
			return nil, err
		}
	}
	// 获取板块内股票映射关系数据
	industryRelationList, err := dal.GetAllStockIndustryRelation(ctx)
	if err != nil {
		return nil, err
	}
	industryRelationMap := make(map[string][]string)
	for _, item := range industryRelationList {
		industryRelationMap[item.IndustryCode] = append(industryRelationMap[item.IndustryCode], item.CompanyCode)
	}
	// 获取股票数据
	stockCodeList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}
	stockCodeMap := make(map[string]string)
	for _, stockCode := range stockCodeList {
		stockCodeMap[stockCode.CompanyCode] = stockCode.CompanyName
	}
	// 计算返回结果
	ret := make([]*model.IndustryBasicData, 0, len(industryList))
	for _, industry := range industryList {
		companyCodeList := industryRelationMap[industry.Code]
		if len(companyCodeList) == 0 {
			continue
		}

		cml := make([]*model.CodeBasic, 0, len(companyCodeList))
		for _, companyCode := range companyCodeList {
			cml = append(cml, &model.CodeBasic{
				Code: companyCode,
				Name: stockCodeMap[companyCode],
			})
		}

		ret = append(ret, &model.IndustryBasicData{
			IndustryCode:    industry.Code,
			IndustryName:    industry.Name,
			CompanyCodeList: cml,
		})
	}
	return ret, nil
}

func GetIndustryTrendData(ctx context.Context, req *model.GetIndustryTrendDataReq) (*model.GetIndustryTrendDataResp, error) {
	resp := &model.GetIndustryTrendDataResp{}
	if req.IndustryCode == "" {
		trend, err := GetIndustryTrendDetail(ctx, req)
		if err != nil {
			return nil, err
		}
		resp.IndustryPriceTrend = trend
		for _, item := range resp.IndustryPriceTrend {
			for _, priceTrend := range item.PriceTrendList {
				priceTrend.Price = utils.Float64KeepDecimal((priceTrend.Price-1)*100, 2)
			}
		}
	} else {
		trend, err := GetIndustryCodeDetail(ctx, req)
		if err != nil {
			return nil, err
		}
		resp.IndustryCodeTrend = trend
		for _, item := range resp.IndustryCodeTrend {
			for _, priceTrend := range item.PriceTrendList {
				priceTrend.Price = utils.Float64KeepDecimal((priceTrend.Price-1)*100, 2)
			}
		}
	}

	return resp, nil
}

func GetIndustryCodeDetail(ctx context.Context, req *model.GetIndustryTrendDataReq) ([]*model.IndustryCodeTrend, error) {
	// 根据板块获取具体的股票列表
	industryRelationList, err := dal.GetStockIndustryRelation(ctx, req.IndustryCode)
	if err != nil {
		return nil, err
	}
	// 根据股票代码获取详情
	codeList := make([]string, len(industryRelationList))
	for idx, item := range industryRelationList {
		codeList[idx] = item.CompanyCode
	}
	stockCodeList, err := dal.GetStockCodeByCodeList(ctx, codeList)
	if err != nil {
		return nil, err
	}
	stockCodeMap := make(map[string]string)
	for _, stockCode := range stockCodeList {
		stockCodeMap[stockCode.CompanyCode] = stockCode.CompanyName
	}
	// 获取股价数据
	stockPriceMap, err := getStockPrice(ctx, codeList, req)
	if err != nil {
		return nil, err
	}
	// 过滤掉时间不符合的股票价格
	stockPriceMap = filterStockPrice(stockPriceMap)
	// 处理股票价格数据
	codeDateMap := make(map[string][]*model.CodeDiffPrice)
	for stockCode, stockPriceList := range stockPriceMap {
		if _, ok := codeDateMap[stockCode]; !ok {
			codeDateMap[stockCode] = make([]*model.CodeDiffPrice, 0)
		}
		// 这里是倒序的,第一个是最新的价格
		lastStockPrice := stockPriceList[len(stockPriceList)-1]
		for i := 0; i < len(stockPriceList)-1; i++ {
			stockPrice := stockPriceList[i]
			nextStockPrice := stockPriceList[i+1]
			date := utils.FormatDate(stockPrice.Date)
			diff := utils.Float64KeepDecimal(100*(stockPrice.PriceClose-nextStockPrice.PriceClose)/nextStockPrice.PriceClose, 2)
			price := utils.Float64KeepDecimal(100*(stockPrice.PriceClose-lastStockPrice.PriceClose)/lastStockPrice.PriceClose, 4)
			codeDateMap[stockCode] = append(codeDateMap[stockCode], &model.CodeDiffPrice{
				Date:  date,
				Diff:  diff,
				Price: price,
				Code:  stockCode,
			})
		}

		// 计算资金流入数据, 这里采用折中的方案, 如果几乎所有的股票都没有这个数据的话, 则不予计算
		inflowMap := make(map[string]*model.FundInflowItem)
		var mainInflowAmount int64 = 0
		var extremeLargeInflowAmount int64 = 0
		var largeInflowAmount int64 = 0
		var mediumInflowAmount int64 = 0
		var smallInflowAmount int64 = 0
		for i := len(stockPriceList) - 2; i >= 0; i-- {
			stockPrice := stockPriceList[i]
			mainInflowAmount += stockPrice.MainInflowAmount
			extremeLargeInflowAmount += stockPrice.ExtremeLargeInflowAmount
			largeInflowAmount += stockPrice.LargeInflowAmount
			mediumInflowAmount += stockPrice.MediumInflowAmount
			smallInflowAmount += stockPrice.SmallInflowAmount
			date := utils.FormatDate(stockPrice.Date)
			idx := fmt.Sprintf("%s_%s", stockCode, date)
			inflowMap[idx] = &model.FundInflowItem{
				MainInflowAmount:         mainInflowAmount,
				ExtremeLargeInflowAmount: extremeLargeInflowAmount,
				LargeInflowAmount:        largeInflowAmount,
				MediumInflowAmount:       mediumInflowAmount,
				SmallInflowAmount:        smallInflowAmount,
			}
		}
		for _, codeDiff := range codeDateMap[stockCode] {
			idx := fmt.Sprintf("%s_%s", codeDiff.Code, codeDiff.Date)
			if item, ok := inflowMap[idx]; ok {
				codeDiff.MainInflowAmount = item.MainInflowAmount
				codeDiff.ExtremeLargeInflowAmount = item.ExtremeLargeInflowAmount
				codeDiff.LargeInflowAmount = item.LargeInflowAmount
				codeDiff.MediumInflowAmount = item.MediumInflowAmount
				codeDiff.SmallInflowAmount = item.SmallInflowAmount
			}
		}
	}

	ret := make([]*model.IndustryCodeTrend, 0)
	for stockCode, diffList := range codeDateMap {
		d := &model.IndustryCodeTrend{
			StockCode:      stockCode,
			StockName:      stockCodeMap[stockCode],
			PriceTrendList: make([]*model.PriceTrend, 0),
		}
		for _, p := range diffList {
			d.PriceTrendList = append(d.PriceTrendList, &model.PriceTrend{
				DateString: p.Date,
				Diff:       utils.Float64KeepDecimal(p.Diff, 2),
				Price:      utils.Float64KeepDecimal((100+p.Price)/100, 4),
				Date:       utils.ParseDate(p.Date),
				FundInflowItem: model.FundInflowItem{
					MainInflowAmount:         p.MainInflowAmount,
					ExtremeLargeInflowAmount: p.ExtremeLargeInflowAmount,
					LargeInflowAmount:        p.LargeInflowAmount,
					MediumInflowAmount:       p.MediumInflowAmount,
					SmallInflowAmount:        p.SmallInflowAmount,
				},
			})
		}
		sort.Sort(model.SortPriceTrend(d.PriceTrendList))
		ret = append(ret, d)
	}
	sort.Sort(model.SortIndustryCodeTrend(ret))
	return ret, nil
}

func WrapGetIndustryTrendDetail(ctx context.Context, req *model.GetIndustryTrendDataReq, industryList []*dal.StockIndustry) ([]*model.IndustryPriceTrend, error) {
	industryDalMap := make(map[string]*dal.StockIndustry)
	for _, industry := range industryList {
		industryDalMap[industry.Code] = industry
	}

	stockCodeList := make([]string, 0)
	industryMap := make(map[string]string)
	for _, industry := range industryList {
		industryRelationList, err := dal.GetStockIndustryRelation(ctx, industry.Code)
		if err != nil {
			return nil, err
		}
		for _, item := range industryRelationList {
			stockCodeList = append(stockCodeList, item.CompanyCode)
			industryMap[item.CompanyCode] = industry.Code
		}
	}
	// 为了避免同步股价的数据导致接口响应过慢，先检查最新的股价数据是否存在，如果不存在就同步。
	stockPriceMap, err := getStockPrice(ctx, stockCodeList, req)
	if err != nil {
		return nil, err
	}

	// 过滤掉时间不符合的股票价格
	stockPriceMap = filterStockPrice(stockPriceMap)

	industryDateMap := make(map[string][]*model.CodeDiffPrice)
	for stockCode, stockPriceList := range stockPriceMap {
		if len(stockPriceList) == 0 {
			continue
		}
		industryCode := industryMap[stockCode]
		if _, ok := industryDateMap[industryCode]; !ok {
			industryDateMap[industryCode] = make([]*model.CodeDiffPrice, 0)
		}
		// 这里是倒序的,第一个是最新的价格
		lastStockPrice := stockPriceList[len(stockPriceList)-1]
		for i := 0; i < len(stockPriceList)-1; i++ {
			stockPrice := stockPriceList[i]
			nextStockPrice := stockPriceList[i+1]
			date := utils.FormatDate(stockPrice.Date)
			diff := utils.Float64KeepDecimal(100*(stockPrice.PriceClose-nextStockPrice.PriceClose)/nextStockPrice.PriceClose, 4)
			price := utils.Float64KeepDecimal(100*(stockPrice.PriceClose-lastStockPrice.PriceClose)/lastStockPrice.PriceClose, 4)
			industryDateMap[industryCode] = append(industryDateMap[industryCode], &model.CodeDiffPrice{
				Date:  date,
				Diff:  diff,
				Price: price,
				Code:  stockCode,
			})
		}
		// 计算资金流入数据, 这里采用折中的方案, 如果几乎所有的股票都没有这个数据的话, 则不予计算
		inflowMap := make(map[string]*model.FundInflowItem)
		var mainInflowAmount int64 = 0
		var extremeLargeInflowAmount int64 = 0
		var largeInflowAmount int64 = 0
		var mediumInflowAmount int64 = 0
		var smallInflowAmount int64 = 0
		for i := len(stockPriceList) - 2; i >= 0; i-- {
			stockPrice := stockPriceList[i]
			mainInflowAmount += stockPrice.MainInflowAmount
			extremeLargeInflowAmount += stockPrice.ExtremeLargeInflowAmount
			largeInflowAmount += stockPrice.LargeInflowAmount
			mediumInflowAmount += stockPrice.MediumInflowAmount
			smallInflowAmount += stockPrice.SmallInflowAmount
			date := utils.FormatDate(stockPrice.Date)
			idx := fmt.Sprintf("%s_%s", stockCode, date)
			inflowMap[idx] = &model.FundInflowItem{
				MainInflowAmount:         mainInflowAmount,
				ExtremeLargeInflowAmount: extremeLargeInflowAmount,
				LargeInflowAmount:        largeInflowAmount,
				MediumInflowAmount:       mediumInflowAmount,
				SmallInflowAmount:        smallInflowAmount,
			}
		}
		for _, codeDiff := range industryDateMap[industryCode] {
			idx := fmt.Sprintf("%s_%s", codeDiff.Code, codeDiff.Date)
			if item, ok := inflowMap[idx]; ok {
				codeDiff.MainInflowAmount = item.MainInflowAmount
				codeDiff.ExtremeLargeInflowAmount = item.ExtremeLargeInflowAmount
				codeDiff.LargeInflowAmount = item.LargeInflowAmount
				codeDiff.MediumInflowAmount = item.MediumInflowAmount
				codeDiff.SmallInflowAmount = item.SmallInflowAmount
			}
		}
	}

	ret := make([]*model.IndustryPriceTrend, 0)
	for industryCode, diffList := range industryDateMap {
		industry := industryDalMap[industryCode]
		d := &model.IndustryPriceTrend{
			IndustryCode:   industry.Code,
			IndustryName:   industry.Name,
			PriceTrendList: make([]*model.PriceTrend, 0),
		}
		diffMap := make(map[string][]float64)
		priceMap := make(map[string][]float64)
		mainInflowMap := make(map[string][]int64)
		extremeLargeInflowMap := make(map[string][]int64)
		largeInflowMap := make(map[string][]int64)
		mediumInflowMap := make(map[string][]int64)
		smallInflowMap := make(map[string][]int64)
		for _, p := range diffList {
			diffMap[p.Date] = append(diffMap[p.Date], p.Diff)
			priceMap[p.Date] = append(priceMap[p.Date], (100+p.Price)/100)
			mainInflowMap[p.Date] = append(mainInflowMap[p.Date], p.MainInflowAmount)
			extremeLargeInflowMap[p.Date] = append(extremeLargeInflowMap[p.Date], p.ExtremeLargeInflowAmount)
			largeInflowMap[p.Date] = append(largeInflowMap[p.Date], p.LargeInflowAmount)
			mediumInflowMap[p.Date] = append(mediumInflowMap[p.Date], p.MediumInflowAmount)
			smallInflowMap[p.Date] = append(smallInflowMap[p.Date], p.SmallInflowAmount)
		}
		for date, dl := range diffMap {
			d.PriceTrendList = append(d.PriceTrendList, &model.PriceTrend{
				DateString: date,
				Diff:       utils.Float64KeepDecimal(utils.ListFloat64Average(dl), 4),
				Price:      utils.Float64KeepDecimal(utils.ListFloat64Average(priceMap[date]), 4),
				Date:       utils.ParseDate(date),
				FundInflowItem: model.FundInflowItem{
					MainInflowAmount:         utils.ListSum(mainInflowMap[date]),
					ExtremeLargeInflowAmount: utils.ListSum(extremeLargeInflowMap[date]),
					LargeInflowAmount:        utils.ListSum(largeInflowMap[date]),
					MediumInflowAmount:       utils.ListSum(mediumInflowMap[date]),
					SmallInflowAmount:        utils.ListSum(smallInflowMap[date]),
				},
			})
		}
		sort.Sort(model.SortPriceTrend(d.PriceTrendList))
		ret = append(ret, d)
	}
	sort.Sort(model.SortIndustryPriceTrend(ret))
	return ret, nil
}

// 获取板块的走势图
func GetIndustryTrendDetail(ctx context.Context, req *model.GetIndustryTrendDataReq) ([]*model.IndustryPriceTrend, error) {
	// 根据板块内的股票的波动率，计算当天板块的波动率
	industryList, err := dal.GetAllStockIndustry(ctx)
	if err != nil {
		return nil, err
	}
	return WrapGetIndustryTrendDetail(ctx, req, industryList)
}

func GetIndustryTrendDetailByIndustryCode(ctx context.Context, req *model.GetIndustryTrendDataReq, industryCode string) ([]*model.IndustryPriceTrend, error) {
	industry, err := dal.GetStockIndustry(ctx, industryCode)
	if err != nil {
		return nil, err
	}
	industryList := []*dal.StockIndustry{industry}
	return WrapGetIndustryTrendDetail(ctx, req, industryList)
}

func filterStockPrice(stockPriceMap map[string][]*dal.StockPrice) map[string][]*dal.StockPrice {
	var maxDate time.Time
	var minDate time.Time
	ret := make(map[string][]*dal.StockPrice)
	maxDataMap := make(map[time.Time]int)
	maxCount := 0
	minDateMap := make(map[time.Time]int)
	minCount := 0
	for _, stockPriceList := range stockPriceMap {
		if len(stockPriceList) == 0 {
			continue
		}
		latestDate := stockPriceList[0].Date
		if _, ok := maxDataMap[latestDate]; !ok {
			maxDataMap[latestDate] = 0
		}
		maxDataMap[latestDate]++
		if maxDataMap[latestDate] > maxCount {
			maxCount = maxDataMap[latestDate]
			maxDate = latestDate
		}

		latestDate = stockPriceList[len(stockPriceList)-1].Date
		if _, ok := minDateMap[latestDate]; !ok {
			minDateMap[latestDate] = 0
		}
		minDateMap[latestDate]++
		if minDateMap[latestDate] > minCount {
			minCount = minDateMap[latestDate]
			minDate = latestDate
		}
	}

	dateList := make([]string, 0)
	for _, stockPriceList := range stockPriceMap {
		if len(stockPriceList) == 0 {
			continue
		}
		if stockPriceList[0].Date == maxDate && stockPriceList[len(stockPriceList)-1].Date == minDate {
			for _, stockPrice := range stockPriceList {
				date := utils.FormatDate(stockPrice.Date)
				dateList = append(dateList, date)
			}
			break
		}
	}
	for stockCode, stockPriceList := range stockPriceMap {
		// 新股会出现之前的日期数量不够的情况, 这里不计算, 因为缺少上市的价格数据
		if len(stockPriceList) == 0 || len(stockPriceList) < len(dateList) {
			continue
		}
		tmpStockPriceList := make([]*dal.StockPrice, 0, len(dateList))
		dateIdx := 0
		for idx := 0; idx < len(dateList); idx++ {
			if utils.FormatDate(stockPriceList[dateIdx].Date) == dateList[idx] {
				tmpStockPriceList = append(tmpStockPriceList, stockPriceList[dateIdx])
				dateIdx++
			} else {
				tmpStockPriceList = append(tmpStockPriceList, &dal.StockPrice{
					Date:       utils.ParseDate(dateList[idx]),
					PriceClose: stockPriceList[dateIdx].PriceClose,
				})
			}
		}
		ret[stockCode] = tmpStockPriceList
	}
	return ret
}

func syncStockIndustryCode(ctx context.Context, stockCodeList []string) error {
	jobCh := make(chan struct{}, MaxJobNum)
	wg := sync.WaitGroup{}
	canceled := false
	for _, stockCode := range stockCodeList {
		wg.Add(1)
		go func(stockCode string) {
			defer wg.Done()
			defer func() {
				<-jobCh
			}()
			jobCh <- struct{}{}
			if canceled {
				return
			}
			err := syncOneStockCode(ctx, &model.SyncStockCodeReq{
				Code: stockCode,
			})
			if err != nil {
				log.Printf("sync stock code failed: %v", err)
				canceled = true
			}
		}(stockCode)
	}
	wg.Wait()
	return nil
}

func getStockPrice(ctx context.Context, stockCodeList []string, req *model.GetIndustryTrendDataReq) (map[string][]*dal.StockPrice, error) {
	var lastDate string
	var remoteLastDate string
	dateOfToday := utils.GetDateOfToday()
	closeTime := fmt.Sprintf("%s 16:00:00", dateOfToday)
	if time.Now().After(utils.ParseTime(closeTime)) {
		lastDate = dateOfToday
	} else {
		lastDate = utils.GetDateOfLastDays(1)
	}

	if req.SyncPrice {
		client := NewEastMoneyClient()
		stockDailyData, err := client.GetRemoteStockDaily(ctx, stockCodeList[0], utils.ParseDate(lastDate))
		if err != nil {
			return nil, err
		}
		if stockDailyData != nil && len(stockDailyData.Item) > 0 {
			item := stockDailyData.Item[len(stockDailyData.Item)-1]
			timestampIndex := stockDailyData.GetColumnIndexByKey("timestamp")
			timestamp, _ := strconv.ParseInt(utils.ToString(item[timestampIndex]), 10, 64)
			remoteLastDate = utils.TimestampToDate(timestamp / int64(time.Microsecond))
		}

		needSyncCodeList := make([]string, 0)
		jobList := make([]func() (interface{}, error), 0)
		for _, stockCode := range stockCodeList {
			jobList = append(jobList, func(stockCode string) func() (interface{}, error) {
				return func() (interface{}, error) {
					d, err := dal.GetLastStockPrice(ctx, stockCode)
					return &WrapJob{
						Data:      d,
						StockCode: stockCode,
					}, err
				}
			}(stockCode))
		}
		stockPriceList, err := utils.ConcurrentActuator(jobList, MaxDBJobNum)
		if err != nil {
			return nil, err
		}
		for _, stockPrice := range stockPriceList {
			price := stockPrice.(*WrapJob)
			if price.Data != nil {
				lastPriceDate := utils.FormatDate(price.Data.Date)
				if lastPriceDate != remoteLastDate && price.Data.UpdateTime.Before(time.Now().AddDate(0, 0, -1)) {
					needSyncCodeList = append(needSyncCodeList, price.StockCode)
				}
			} else {
				needSyncCodeList = append(needSyncCodeList, price.StockCode)
			}
		}

		// 需要同步所有的数据
		err = syncStockIndustryCode(ctx, needSyncCodeList)
		if err != nil {
			return nil, err
		}
	}

	// 返回对应天数的股价数据
	ret := make(map[string][]*dal.StockPrice, 0)
	jobList2 := make([]func() (interface{}, error), 0)
	for _, stockCode := range stockCodeList {
		jobList2 = append(jobList2, func(stockCode string) func() (interface{}, error) {
			return func() (interface{}, error) {
				d, err := dal.GetLastNStockPrice(ctx, stockCode, req.EndDate, req.Days+1)
				return &WrapJobs{
					Data:      d,
					StockCode: stockCode,
				}, err
			}
		}(stockCode))
	}
	stockPriceList2, err := utils.ConcurrentActuator(jobList2, MaxDBJobNum)
	if err != nil {
		return nil, err
	}
	for _, stockPrice := range stockPriceList2 {
		price := stockPrice.(*WrapJobs)
		ret[price.StockCode] = price.Data
	}
	return ret, nil
}

func GetIndustryRelationData(ctx context.Context, req *model.GetIndustryRelationDataReq) (*model.GetIndustryRelationDataResp, error) {
	if req.IsSplitIndustry {
		return GetIndustryRelationBySplitIndustry(ctx, req)
	} else {
		return GetIndustryRelationByBasicCode(ctx, req)
	}
}

func GetIndustryRelationByBasicCode(ctx context.Context, req *model.GetIndustryRelationDataReq) (*model.GetIndustryRelationDataResp, error) {
	// 获取股价趋势
	trendReq := &model.GetIndustryTrendDataReq{
		Days:         req.Days,
		SyncPrice:    false,
		IndustryCode: req.IndustryCode,
	}
	trendList, err := GetIndustryTrendDetail(ctx, trendReq)
	if err != nil {
		return nil, err
	}
	// 获取上证指数, 计算皮尔逊系数
	stockDailyData, err := GetBasicStockPrice(ctx)
	if err != nil {
		return nil, err
	}
	// 计算相关的皮尔逊系数
	ret := make([]*model.IndustryRelation, 0)
	parsedStockDailyData := stockDailyData.ToDatePriceList()
	stockDailyTrendData := model.TransferDatePriceToPriceTrend(parsedStockDailyData)
	for _, trend := range trendList {
		// 计算相关的皮尔逊系数
		correlation := CalculatePearsonCorrelation(trend.PriceTrendList, stockDailyTrendData)
		ret = append(ret, &model.IndustryRelation{
			IndustryCode:      trend.IndustryCode,
			IndustryName:      trend.IndustryName,
			Correlation:       utils.Float64KeepDecimal(correlation, 2),
			CorrelationString: utils.GetCorrelationString(correlation),
		})
	}
	sort.Sort(model.SortIndustryRelation(ret))
	return &model.GetIndustryRelationDataResp{
		StartDate:            trendList[0].PriceTrendList[0].DateString,
		EndDate:              trendList[len(trendList)-1].PriceTrendList[len(trendList[len(trendList)-1].PriceTrendList)-1].DateString,
		IndustryRelationList: ret,
	}, nil
}

func GetIndustryRelationBySplitIndustry(ctx context.Context, req *model.GetIndustryRelationDataReq) (*model.GetIndustryRelationDataResp, error) {
	// 获取股价趋势
	trendReq := &model.GetIndustryTrendDataReq{
		Days:         req.Days,
		SyncPrice:    false,
		IndustryCode: req.IndustryCode,
	}
	trendList, err := GetIndustryTrendDetail(ctx, trendReq)
	if err != nil {
		return nil, err
	}
	// 计算相关的皮尔逊系数
	ret := make([][]*model.IndustryRelation, 0)
	for {
		strongIndustryRelationList, weakPriceTrendList := GetOneStrongCorrelationList(trendList)
		sort.Sort(model.SortIndustryRelation(strongIndustryRelationList))
		ret = append(ret, strongIndustryRelationList)
		if len(weakPriceTrendList) == 0 {
			break
		}
		trendList = weakPriceTrendList
	}
	return &model.GetIndustryRelationDataResp{
		StartDate:                 trendList[0].PriceTrendList[0].DateString,
		EndDate:                   trendList[len(trendList)-1].PriceTrendList[len(trendList[len(trendList)-1].PriceTrendList)-1].DateString,
		SplitIndustryRelationList: ret,
	}, nil
}

func GetOneStrongCorrelationList(trend []*model.IndustryPriceTrend) ([]*model.IndustryRelation, []*model.IndustryPriceTrend) {
	if len(trend) == 0 {
		return nil, nil
	}
	strongIndustryRelationList := make([]*model.IndustryRelation, 0)
	weakPriceTrendList := make([]*model.IndustryPriceTrend, 0)
	strongIndustryRelationList = append(strongIndustryRelationList, &model.IndustryRelation{
		IndustryCode:      trend[0].IndustryCode,
		IndustryName:      trend[0].IndustryName,
		Correlation:       1.0,
		CorrelationString: "基准",
	})
	for idx := 1; idx < len(trend); idx++ {
		correlation := CalculatePearsonCorrelation(trend[idx].PriceTrendList, trend[0].PriceTrendList)
		if utils.IsStrongCorrelation(correlation) {
			strongIndustryRelationList = append(strongIndustryRelationList, &model.IndustryRelation{
				IndustryCode:      trend[idx].IndustryCode,
				IndustryName:      trend[idx].IndustryName,
				Correlation:       utils.Float64KeepDecimal(correlation, 2),
				CorrelationString: "强相关",
			})
		} else {
			weakPriceTrendList = append(weakPriceTrendList, trend[idx])
		}
	}
	return strongIndustryRelationList, weakPriceTrendList
}

func GetBasicStockPrice(ctx context.Context) (*model.StockDailyData, error) {
	// 默认获取上证指数
	stockCode := utils.GetBasicStockCode()
	client := NewEastMoneyClient()
	stockDailyData, err := client.GetRemoteStockDaily(ctx, stockCode, time.Now())
	if err != nil {
		return nil, err
	}
	return stockDailyData, nil
}

func CalculatePearsonCorrelation(trend1, trend2 []*model.PriceTrend) float64 {
	// 计算相关的皮尔逊系数
	x, y := alignPriceTrend(trend1, trend2)
	// 计算均值
	var sumX, sumY float64
	for i := range trend1 {
		sumX += x[i].Diff
		sumY += y[i].Diff
	}
	meanX := sumX / float64(len(x))
	meanY := sumY / float64(len(y))

	// 计算协方差和标准差
	var cov, stdX, stdY float64
	for i := range x {
		devX := x[i].Diff - meanX
		devY := y[i].Diff - meanY
		cov += devX * devY
		stdX += devX * devX
		stdY += devY * devY
	}

	stdX = math.Sqrt(stdX)
	stdY = math.Sqrt(stdY)

	if stdX == 0 || stdY == 0 {
		return math.NaN()
	}

	return cov / (stdX * stdY)
}

func alignPriceTrend(trend1, trend2 []*model.PriceTrend) ([]*model.PriceTrend, []*model.PriceTrend) {
	// 两者长度不等, 需要根据日期对其以下长度
	longest := trend1
	shortest := trend2
	if len(trend1) < len(trend2) {
		longest = trend2
		shortest = trend1
	}
	for idx, item := range longest {
		if item.DateString == shortest[0].DateString {
			longest = longest[idx:]
			break
		}
	}
	if len(longest) == len(shortest) {
		return longest, shortest
	}
	// 长度不相等, 取最短的长度
	longest = longest[:len(shortest)]
	return longest, shortest
}
