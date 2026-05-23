package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	ConceptCacheKey  = "concept_key"
	MaxConceptJobNum = 3
)

// GetConcepts 获取所有概念列表
func GetConcepts(ctx context.Context, req *model.GetConceptsReq) (*model.GetConceptsResp, error) {
	concepts, err := dal.GetConcepts(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.ConceptResp, 0, len(concepts))
	for _, concept := range concepts {
		item := &model.ConceptResp{
			ID:   concept.ID,
			Name: concept.Name,
		}

		// 解析股票列表（逗号分隔）
		if concept.Stocks != "" {
			stockCodes := strings.Split(concept.Stocks, ",")
			stocks := make([]*model.ConceptStockInfo, 0, len(stockCodes))
			for _, code := range stockCodes {
				code = strings.TrimSpace(code)
				if code == "" {
					continue
				}
				stock, err := dal.GetStockCodeByCode(ctx, code)
				if err != nil {
					return nil, err
				}
				if stock != nil {
					stocks = append(stocks, &model.ConceptStockInfo{
						Code:       stock.CompanyCode,
						Name:       stock.CompanyName,
						MaxPercent: GetLimitUpMaxPercent(utils.GetStockCodeNumber(stock.CompanyCode)),
					})
				} else {
					stocks = append(stocks, &model.ConceptStockInfo{
						Code:       code,
						Name:       "未知",
						MaxPercent: GetLimitUpMaxPercent(utils.GetStockCodeNumber(code)),
					})
				}
			}
			item.Stocks = stocks
		} else {
			item.Stocks = make([]*model.ConceptStockInfo, 0)
		}

		result = append(result, item)
	}

	resp := &model.GetConceptsResp{}
	// 填充涨跌幅数据
	if req.WithChange {
		// 检查缓存, 如果缓存里面有数据的话, 则直接get缓存数据
		cache := GetMemCache(ConceptCacheKey)
		if cache != nil && !req.ForceSync {
			d := cache.Data.([]*model.ConceptResp)
			dMap := make(map[uint]*model.ConceptResp)
			// 防止用户添加了新的概念和股票, 而数据展示不出来
			for _, item := range d {
				dMap[item.ID] = item
			}
			// 合并数据
			for _, item := range result {
				dStockMap := make(map[string]*model.ConceptStockInfo)
				if dItem, ok := dMap[item.ID]; ok {
					item.Percent = dItem.Percent
					for _, stock := range dItem.Stocks {
						dStockMap[stock.Code] = stock
					}
				}

				for _, stock := range item.Stocks {
					if dStock, ok := dStockMap[stock.Code]; ok {
						stock.Percent = dStock.Percent
					}
				}
			}
			// 按涨跌幅排序
			sort.Sort(model.ConceptRespChangeSorter(result))
			resp.GetPercentTime = utils.FormatTime(cache.SetTime)
		} else {
			// 九点半之前的话, 则不填充涨跌幅数据
			if utils.IsBeforeHourMinute(9, 30) {
				sort.Sort(model.ConceptRespNameSorter(result))

				return &model.GetConceptsResp{
					Concepts:       result,
					GetPercentTime: utils.FormatTime(time.Now()),
				}, nil
			}
			// 调用接口获取数据并设置内存缓存
			jobList := make([]func() (interface{}, error), 0)
			for _, concept := range result {
				for _, stock := range concept.Stocks {
					jobList = append(jobList, func(code string) func() (interface{}, error) {
						return func() (interface{}, error) {
							client := NewBaiduClient()
							priceList, err := client.GetRemoteStockMinute(ctx, code)
							if err != nil {
								return nil, err
							}
							if len(priceList) == 0 {
								return nil, nil
							}
							return &model.WrapMinutePrice{
								Code:            code,
								StockMinuteData: priceList[len(priceList)-1],
							}, nil
						}
					}(stock.Code))
				}
			}
			// 执行并发任务
			dataList, err := utils.ConcurrentActuator(jobList, MaxConceptJobNum)
			if err != nil {
				return nil, err
			}
			dataMap := make(map[string]*model.WrapMinutePrice)
			for _, item := range dataList {
				if item == nil {
					continue
				}
				wrap := item.(*model.WrapMinutePrice)
				dataMap[wrap.Code] = wrap
			}
			for _, concept := range result {
				// 填充涨跌幅数据
				concept.Percent = 0.0
				count := 0
				for _, stock := range concept.Stocks {
					if wrap, ok := dataMap[stock.Code]; ok {
						stock.Percent = wrap.Percent
						concept.Percent += wrap.Percent
						count++
					}
				}
				if count > 0 {
					concept.Percent /= float64(count)
					concept.Percent = utils.Float64KeepDecimal(concept.Percent, 2)
				}
			}
			// 按涨跌幅排序
			sort.Sort(model.ConceptRespChangeSorter(result))

			// 设置缓存
			if utils.IsBeforeHourMinute(15, 0) {
				SetMemCache(ConceptCacheKey, result, 3*time.Minute)
			} else {
				SetMemCache(ConceptCacheKey, result, 10*time.Hour)
			}
			resp.GetPercentTime = utils.FormatTime(time.Now())
		}
	} else {
		// 按名称排序
		sort.Sort(model.ConceptRespNameSorter(result))
		resp.GetPercentTime = ""
	}

	resp.Concepts = result
	return resp, nil
}

// AddConcept 添加新概念
func AddConcept(ctx context.Context, req *model.AddConceptReq) error {
	if req.Name == "" {
		return errors.New("concept name is empty")
	}

	// 检查是否已存在同名概念
	exist, err := dal.IsConceptExist(ctx, req.Name)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("concept already exists")
	}

	concept := &dal.Concept{
		Name:   req.Name,
		Stocks: "",
	}

	return dal.CreateConcept(ctx, concept)
}

// DeleteConcept 删除概念
func DeleteConcept(ctx context.Context, req *model.DeleteConceptReq) error {
	if req.ID <= 0 {
		return errors.New("invalid concept id")
	}

	return dal.DeleteConcept(ctx, uint(req.ID))
}

// GetConceptStocks 获取概念下的股票列表
func GetConceptStocks(ctx context.Context, req *model.GetConceptStocksReq) ([]*model.ConceptStockInfo, error) {
	if req.ConceptID <= 0 {
		return nil, errors.New("invalid concept id")
	}

	concept, err := dal.GetConcept(ctx, uint(req.ConceptID))
	if err != nil {
		return nil, err
	}
	if concept == nil {
		return make([]*model.ConceptStockInfo, 0), nil
	}

	result := make([]*model.ConceptStockInfo, 0)
	if concept.Stocks != "" {
		stockCodes := strings.Split(concept.Stocks, ",")
		for _, code := range stockCodes {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			stock, err := dal.GetStockCodeByCode(ctx, code)
			if err != nil {
				return nil, err
			}
			if stock != nil {
				result = append(result, &model.ConceptStockInfo{
					Code: stock.CompanyCode,
					Name: stock.CompanyName,
				})
			} else {
				result = append(result, &model.ConceptStockInfo{
					Code: code,
					Name: "未知",
				})
			}
		}
	}

	return result, nil
}

// AddConceptStock 向概念添加股票
func AddConceptStock(ctx context.Context, req *model.AddConceptStockReq) error {
	if req.ConceptID <= 0 {
		return errors.New("invalid concept id")
	}
	if req.StockCode == "" {
		return errors.New("stock code is empty")
	}

	// 验证股票代码
	stockCode := req.StockCode
	if utils.IsStockNumber(stockCode) {
		stockCode = utils.GetFullStockCodeOfNumber(stockCode)
	} else if utils.IsStockCodeWithPrefix(stockCode) {
		stockCode = strings.ToUpper(stockCode)
	} else {
		return errors.New("invalid stock code format")
	}

	// 检查股票是否存在
	exist, err := dal.IsStockCodeExist(ctx, stockCode)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("stock code not found")
	}

	// 获取概念
	concept, err := dal.GetConcept(ctx, uint(req.ConceptID))
	if err != nil {
		return err
	}
	if concept == nil {
		return errors.New("concept not found")
	}

	// 检查股票是否已存在
	if concept.Stocks != "" {
		stockCodes := strings.Split(concept.Stocks, ",")
		for _, code := range stockCodes {
			if strings.TrimSpace(code) == stockCode {
				return errors.New("stock already exists in concept")
			}
		}
		// 添加股票
		concept.Stocks = concept.Stocks + "," + stockCode
	} else {
		concept.Stocks = stockCode
	}

	return dal.UpdateConcept(ctx, concept)
}

// DeleteConceptStock 从概念中移除股票
func DeleteConceptStock(ctx context.Context, req *model.DeleteConceptStockReq) error {
	if req.ConceptID <= 0 {
		return errors.New("invalid concept id")
	}
	if req.StockCode == "" {
		return errors.New("stock code is empty")
	}

	// 获取概念
	concept, err := dal.GetConcept(ctx, uint(req.ConceptID))
	if err != nil {
		return err
	}
	if concept == nil {
		return errors.New("concept not found")
	}

	if concept.Stocks == "" {
		return errors.New("no stocks in concept")
	}

	// 移除股票
	stockCodes := strings.Split(concept.Stocks, ",")
	newStocks := make([]string, 0, len(stockCodes))
	for _, code := range stockCodes {
		if strings.TrimSpace(code) != req.StockCode {
			newStocks = append(newStocks, strings.TrimSpace(code))
		}
	}

	concept.Stocks = strings.Join(newStocks, ",")

	return dal.UpdateConcept(ctx, concept)
}
