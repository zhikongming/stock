package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/utils"
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
	limit := 5
	stockPriceList, err := dal.GetLastNStockPrice(ctx, companyCode, "", limit)
	if err != nil {
		return nil, err
	}
	if len(stockPriceList) != limit {
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
