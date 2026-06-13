package service

import (
	"context"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
)

func GetCodeBasicByCodeList(ctx context.Context, codeList []string) ([]*model.CodeBasic, error) {
	if len(codeList) == 0 {
		return make([]*model.CodeBasic, 0), nil
	}
	stockCodeList, err := dal.GetStockCodeByCodeList(ctx, codeList)
	if err != nil {
		return nil, err
	}
	result := make([]*model.CodeBasic, 0, len(stockCodeList))
	for _, stock := range stockCodeList {
		result = append(result, &model.CodeBasic{
			Code: stock.CompanyCode,
			Name: stock.CompanyName,
		})
	}
	return result, nil
}
