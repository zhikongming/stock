package service

import (
	"context"
	"fmt"
	"log"
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
	} else {
		trend, err := GetIndustryCodeDetail(ctx, req)
		if err != nil {
			return nil, err
		}
		resp.IndustryCodeTrend = trend
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
			})
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
			})
		}
		sort.Sort(model.SortPriceTrend(d.PriceTrendList))
		ret = append(ret, d)
	}
	sort.Sort(model.SortIndustryCodeTrend(ret))
	return ret, nil
}

// 获取板块的走势图
func GetIndustryTrendDetail(ctx context.Context, req *model.GetIndustryTrendDataReq) ([]*model.IndustryPriceTrend, error) {
	// 根据板块内的股票的波动率，计算当天板块的波动率
	industryList, err := dal.GetAllStockIndustry(ctx)
	if err != nil {
		return nil, err
	}

	// tmpList := make([]*dal.StockIndustry, 0)
	// for _, item := range industryList {
	// 	if item.Code == "BK0420" {
	// 		tmpList = append(tmpList, item)
	// 	}
	// }
	// industryList = tmpList

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
			diff := utils.Float64KeepDecimal(100*(stockPrice.PriceClose-nextStockPrice.PriceClose)/nextStockPrice.PriceClose, 2)
			price := utils.Float64KeepDecimal(100*(stockPrice.PriceClose-lastStockPrice.PriceClose)/lastStockPrice.PriceClose, 4)
			industryDateMap[industryCode] = append(industryDateMap[industryCode], &model.CodeDiffPrice{
				Date:  date,
				Diff:  diff,
				Price: price,
			})
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
		for _, p := range diffList {
			diffMap[p.Date] = append(diffMap[p.Date], p.Diff)
			priceMap[p.Date] = append(priceMap[p.Date], (100+p.Price)/100)
		}
		for date, dl := range diffMap {
			d.PriceTrendList = append(d.PriceTrendList, &model.PriceTrend{
				DateString: date,
				Diff:       utils.Float64KeepDecimal(utils.ListFloat64Average(dl), 2),
				Price:      utils.Float64KeepDecimal(utils.ListFloat64Average(priceMap[date]), 4),
				Date:       utils.ParseDate(date),
			})
		}
		sort.Sort(model.SortPriceTrend(d.PriceTrendList))
		ret = append(ret, d)
	}
	sort.Sort(model.SortIndustryPriceTrend(ret))
	return ret, nil
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
	for _, stockCode := range stockCodeList {
		wg.Add(1)
		go func(stockCode string) {
			defer wg.Done()
			defer func() {
				<-jobCh
			}()
			jobCh <- struct{}{}
			err := syncOneStockCode(ctx, &model.SyncStockCodeReq{
				Code: stockCode,
			})
			if err != nil {
				log.Printf("sync stock code failed: %v", err)
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
				d, err := dal.GetLastNStockPrice(ctx, stockCode, req.Days+1)
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
