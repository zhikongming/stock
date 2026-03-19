package service

import (
	"context"
	"encoding/json"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
)

// AddPriceAnalyse 添加量价分析股票
func UpdatePriceAnalyse(ctx context.Context, req *model.UpdatePriceAnalyseReq) error {
	if len(req.CodeList) == 0 {
		return nil
	}
	// 这里实现添加量价分析股票的逻辑
	if req.PriceAnalyseType == model.PriceAnalyseTypeAdd {
		return dal.AddParsedPriceCodeList(ctx, req.CodeList)
	} else if req.PriceAnalyseType == model.PriceAnalyseTypeDelete {
		return dal.DeleteParsedPriceCodeList(ctx, req.CodeList)
	}
	return nil
}

// GetPriceAnalyse 获取量价分析结果
func GetPriceAnalyse(ctx context.Context, req *model.GetPriceAnalyseReq) ([]*model.MultiVolumePrice, error) {
	// 如果请求中没有提供股票代码，则从stock_code中获取is_parsed_price=1的数据
	stockCodeList, err := dal.GetStockCodeByParsedPrice(ctx, true)
	if err != nil {
		return nil, err
	}
	codeList := []string{}
	nameList := []string{}
	for _, stockCode := range stockCodeList {
		codeList = append(codeList, stockCode.CompanyCode)
		nameList = append(nameList, stockCode.CompanyName)
	}

	// 尝试从缓存中获取数据
	cozeCache := GetCozeCache()
	cache, err := cozeCache.GetAndSetMultiVolumePrice(ctx, codeList, nameList)
	if err != nil {
		return nil, err
	}

	// 解析缓存数据
	var results []*model.MultiVolumePrice
	err = json.Unmarshal([]byte(cache.DataValue), &results)
	if err != nil {
		return nil, err
	}

	// 填充没有获取到的数据和删除的数据.
	resultMap := make(map[string]*model.MultiVolumePrice)
	for _, item := range results {
		resultMap[item.CompanyCode] = item
	}
	ret := []*model.MultiVolumePrice{}
	for _, stockCode := range stockCodeList {
		item, ok := resultMap[stockCode.CompanyCode]
		if ok {
			ret = append(ret, item)
		} else {
			ret = append(ret, &model.MultiVolumePrice{
				CompanyCode:    stockCode.CompanyCode,
				CompanyName:    stockCode.CompanyName,
				IsSafe:         "-",
				AnalysisResult: "暂无数据",
				StartDate:      "-",
				EndDate:        "-",
			})
		}
	}

	return ret, nil
}
