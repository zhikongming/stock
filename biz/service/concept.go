package service

import (
	"context"
	"errors"
	"strings"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

// GetConcepts 获取所有概念列表
func GetConcepts(ctx context.Context) ([]*model.ConceptResp, error) {
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
						Code: stock.CompanyCode,
						Name: stock.CompanyName,
					})
				} else {
					stocks = append(stocks, &model.ConceptStockInfo{
						Code: code,
						Name: "未知",
					})
				}
			}
			item.Stocks = stocks
		} else {
			item.Stocks = make([]*model.ConceptStockInfo, 0)
		}

		result = append(result, item)
	}

	return result, nil
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
