package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	AnalyzeVolumePriceLimit       = 6
	AnalyzeVolumePriceReportLimit = 10
)

type EMCache struct {
	CookieIndex int
	Timeout     time.Time
	Mutex       sync.Mutex
}

func NewEMCache() *EMCache {
	return &EMCache{
		CookieIndex: 0,
		Timeout:     time.Now().Add(time.Hour * 24 * 356),
	}
}

func (c *EMCache) GetCookieIndex() int {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if time.Now().After(c.Timeout) {
		c.CookieIndex = 0
		c.Timeout = time.Now().Add(time.Hour)
	}
	return c.CookieIndex
}

func (c *EMCache) SetCookieIndex(idx int) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.CookieIndex = idx % len(NidList)
	c.Timeout = time.Now().Add(time.Hour)
}

var emCache = NewEMCache()

type CozeCache struct {
}

var cozeCache *CozeCache

func GetCozeCache() *CozeCache {
	if cozeCache == nil {
		cozeCache = &CozeCache{}
	}
	return cozeCache
}

func (c *CozeCache) GetAndSetSimilarCompany(ctx context.Context, companyCode string, companyName string) (*dal.Cache, error) {
	// 尝试从缓存中获取数据
	cache, err := dal.GetCacheByType(ctx, companyCode, dal.CacheTypeSimilarCompany)
	if err != nil {
		return nil, err
	}
	if cache != nil {
		return cache, nil
	}

	// 调用远端接口, 从接口中获取数据
	client := NewCozeClient()
	if client == nil {
		return nil, errors.New("coze client not found")
	}
	similarCompanies, err := client.GetSimilarCompany(ctx, companyName)
	if err != nil {
		return nil, err
	}
	// 缓存数据
	similarCompaniesByte, _ := json.Marshal(similarCompanies)
	cache = &dal.Cache{
		DataKey:   companyCode,
		DataType:  int8(dal.CacheTypeSimilarCompany),
		Date:      utils.FormatDate(time.Now()),
		DataValue: string(similarCompaniesByte),
	}
	err = dal.CreateCache(ctx, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (c *CozeCache) GetAndSetVolumePrice(ctx context.Context, companyCode string, companyName string) (*dal.Cache, error) {
	// 获取最后5个股价数据
	stockPriceList, err := dal.GetLastNStockPrice(ctx, companyCode, "", AnalyzeVolumePriceLimit)
	if err != nil {
		return nil, err
	}
	if len(stockPriceList) != AnalyzeVolumePriceLimit {
		return nil, errors.New("stock price list not enough")
	}
	// 尝试从缓存中获取数据
	cache, err := dal.GetCacheByTypeDate(ctx, companyCode, dal.CacheTypeVolumePrice, utils.FormatDate(stockPriceList[0].Date))
	if err != nil {
		return nil, err
	}
	if cache != nil {
		return cache, nil
	}

	// 调用远端接口, 从接口中获取数据
	client := NewCozeClient()
	if client == nil {
		return nil, errors.New("coze client not found")
	}
	utils.ListSwap(stockPriceList)
	volumePrice, err := client.GetVolumePrice(ctx, companyName, stockPriceList)
	if err != nil {
		return nil, err
	}
	volumePrice.StartDate = utils.FormatDate(stockPriceList[0].Date)
	volumePrice.EndDate = utils.FormatDate(stockPriceList[len(stockPriceList)-1].Date)
	// 缓存数据
	volumePriceByte, _ := json.Marshal(volumePrice)
	cache = &dal.Cache{
		DataKey:   companyCode,
		DataType:  int8(dal.CacheTypeVolumePrice),
		Date:      utils.FormatDate(stockPriceList[len(stockPriceList)-1].Date),
		DataValue: string(volumePriceByte),
	}
	err = dal.CreateCache(ctx, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (c *CozeCache) GetAndSetBusinessAnalysis(ctx context.Context, companyCode string, companyName string) (*dal.Cache, error) {
	// 尝试从缓存中获取数据
	cache, err := dal.GetCacheByType(ctx, companyCode, dal.CacheTypeBusinessAnalysis)
	if err != nil {
		return nil, err
	}
	if cache != nil {
		// 解析缓存日期
		cacheDate := utils.ParseDate(cache.Date)
		// 如果缓存时间超过一个月，删除缓存并重新获取
		if time.Since(cacheDate) > 30*24*time.Hour {
			err = dal.DeleteCache(ctx, cache.ID)
			if err == nil {
				cache = nil
			}
		}
		if cache != nil {
			return cache, nil
		}
	}

	// 调用远端接口, 从接口中获取数据
	client := NewCozeClient()
	if client == nil {
		return nil, errors.New("coze client not found")
	}
	businessAnalysis, err := client.GetBusinessAnalysis(ctx, companyName)
	if err != nil {
		return nil, err
	}
	// 缓存数据
	businessAnalysisByte, _ := json.Marshal(businessAnalysis)
	cache = &dal.Cache{
		DataKey:   companyCode,
		DataType:  int8(dal.CacheTypeBusinessAnalysis),
		Date:      utils.FormatDate(time.Now()),
		DataValue: string(businessAnalysisByte),
	}
	err = dal.CreateCache(ctx, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (c *CozeCache) GetMultiVolumePrice(ctx context.Context, codeList, nameList []string) ([]*model.MultiVolumePrice, error) {
	// 根据多个股票代码来决定时间
	stockPriceMap := make(map[string][]*dal.StockPrice)
	dateList := make([]string, 0)
	for _, code := range codeList {
		stockPriceList, err := dal.GetLastNStockPrice(ctx, code, "", AnalyzeVolumePriceLimit)
		if err != nil {
			return nil, err
		}
		if len(stockPriceList) != AnalyzeVolumePriceLimit {
			continue
		}
		stockPriceMap[code] = stockPriceList
		dateList = append(dateList, utils.FormatDate(stockPriceList[0].Date))
	}

	// 调用远端接口, 从接口中获取数据
	client := NewCozeClient()
	if client == nil {
		return nil, errors.New("coze client not found")
	}
	params := make([]*model.GetVolumePriceReq, 0)
	for idx, code := range codeList {
		stockPriceList, ok := stockPriceMap[code]
		if !ok {
			continue
		}
		utils.ListSwap(stockPriceList)
		stockDataList := make([]*model.StockData, 0)
		for _, stockPrice := range stockPriceList {
			stockDataList = append(stockDataList, &model.StockData{
				Date:       utils.FormatDate(stockPrice.Date),
				OpenPrice:  utils.ToString(stockPrice.PriceOpen),
				ClosePrice: utils.ToString(stockPrice.PriceClose),
				HighPrice:  utils.ToString(stockPrice.PriceHigh),
				LowPrice:   utils.ToString(stockPrice.PriceLow),
				Volume:     utils.ToString(stockPrice.Amount),
			})
		}
		req := &model.GetVolumePriceReq{
			CompanyName:   nameList[idx],
			StockDataList: stockDataList,
		}
		params = append(params, req)
	}
	resp, err := client.GetMultiVolumePrice(ctx, params)
	if err != nil {
		return nil, err
	}
	// 构建映射数据
	stockMap := make(map[string]string)
	for idx := range nameList {
		stockMap[nameList[idx]] = codeList[idx]
	}
	for _, item := range resp.Results {
		item.CompanyCode = stockMap[item.CompanyName]
		stockPriceList := stockPriceMap[item.CompanyCode]
		item.StartDate = utils.FormatDate(stockPriceList[0].Date)
		item.EndDate = utils.FormatDate(stockPriceList[len(stockPriceList)-1].Date)
	}

	ret := make([]*model.MultiVolumePrice, 0)
	for _, item := range resp.Results {
		if item.IsSafe == IsSafeDirtyStatus {
			var result model.MultiVolumePrice
			sanitized := strings.ReplaceAll(item.AnalysisResult, "\n", "\\n")
			err = json.Unmarshal([]byte(sanitized), &result)
			if err == nil {
				item.IsSafe = result.IsSafe
				item.AnalysisResult = result.AnalysisResult
			}
		}
		ret = append(ret, item)
	}
	return ret, nil
}

func (c *CozeCache) GetAndSetMultiVolumePrice(ctx context.Context, codeList, nameList []string) (*dal.Cache, error) {
	// 根据多个股票代码来决定时间
	stockPriceMap := make(map[string][]*dal.StockPrice)
	dateList := make([]string, 0)
	for _, code := range codeList {
		stockPriceList, err := dal.GetLastNStockPrice(ctx, code, "", AnalyzeVolumePriceLimit)
		if err != nil {
			return nil, err
		}
		if len(stockPriceList) != AnalyzeVolumePriceLimit {
			continue
		}
		stockPriceMap[code] = stockPriceList
		dateList = append(dateList, utils.FormatDate(stockPriceList[0].Date))
	}
	// 取日期列表中出现最多的日期, 尝试从缓存中获取数据
	mostCommonDate := utils.MostCommon(dateList)
	cache, err := dal.GetCacheByTypeDate(ctx, string(dal.CacheKeyMultiVolumePrice), dal.CacheTypeMultiVolumePrice, mostCommonDate)
	if err != nil {
		return nil, err
	}
	if cache != nil {
		return cache, nil
	}

	// 调用远端接口, 从接口中获取数据
	client := NewCozeClient()
	if client == nil {
		return nil, errors.New("coze client not found")
	}
	params := make([]*model.GetVolumePriceReq, 0)
	for idx, code := range codeList {
		stockPriceList, ok := stockPriceMap[code]
		if !ok {
			continue
		}
		utils.ListSwap(stockPriceList)
		stockDataList := make([]*model.StockData, 0)
		for _, stockPrice := range stockPriceList {
			stockDataList = append(stockDataList, &model.StockData{
				Date:       utils.FormatDate(stockPrice.Date),
				OpenPrice:  utils.ToString(stockPrice.PriceOpen),
				ClosePrice: utils.ToString(stockPrice.PriceClose),
				HighPrice:  utils.ToString(stockPrice.PriceHigh),
				LowPrice:   utils.ToString(stockPrice.PriceLow),
				Volume:     utils.ToString(stockPrice.Amount),
			})
		}
		req := &model.GetVolumePriceReq{
			CompanyName:   nameList[idx],
			StockDataList: stockDataList,
		}
		params = append(params, req)
	}
	resp, err := client.GetMultiVolumePrice(ctx, params)
	if err != nil {
		return nil, err
	}
	// 构建映射数据
	stockMap := make(map[string]string)
	for idx := range nameList {
		stockMap[nameList[idx]] = codeList[idx]
	}
	for _, item := range resp.Results {
		item.CompanyCode = stockMap[item.CompanyName]
		stockPriceList := stockPriceMap[item.CompanyCode]
		item.StartDate = utils.FormatDate(stockPriceList[0].Date)
		item.EndDate = utils.FormatDate(stockPriceList[len(stockPriceList)-1].Date)
	}

	// 缓存数据
	volumePriceByte, _ := json.Marshal(resp.Results)
	cache = &dal.Cache{
		DataKey:   string(dal.CacheKeyMultiVolumePrice),
		DataType:  int8(dal.CacheTypeMultiVolumePrice),
		Date:      mostCommonDate,
		DataValue: string(volumePriceByte),
	}
	err = dal.CreateCache(ctx, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (c *CozeCache) GetLastNMultiVolumePrice(ctx context.Context, limit int) ([]*dal.Cache, error) {
	// 从缓存中获取数据
	cache, err := dal.GetCacheByTypeLimit(ctx, string(dal.CacheKeyMultiVolumePrice), dal.CacheTypeMultiVolumePrice, limit)
	if err != nil {
		return nil, err
	}
	return cache, nil
}
