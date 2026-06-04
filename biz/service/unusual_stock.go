package service

import (
	"context"
	"sort"
	"strings"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func TryToAddUnusualStock(ctx context.Context, dataList []*model.UnusualStock) error {
	// 主要是去除重复的数据
	for _, item := range dataList {
		// 去除掉板块
		if !utils.IsStockCodeWithPrefix(item.Code) {
			continue
		}
		if strings.Contains(item.Name, "退市") || strings.Contains(item.Name, "ST") {
			continue
		}
		record, err := dal.GetUnusualStockByCodeTypeEndDate(ctx, item.Code, int(item.Type), item.EndDate)
		if err != nil {
			return err
		}
		if record != nil {
			continue
		}
		if err := dal.CreateUnusualStock(ctx, &dal.UnusualStock{
			Code:          item.Code,
			Name:          item.Name,
			Type:          int(item.Type),
			StartDate:     item.StartDate,
			EndDate:       item.EndDate,
			NoticeDate:    item.NoticeDate,
			UnusualType:   item.UnusualType,
			UnusualReason: item.UnusualReason,
		}); err != nil {
			return err
		}
	}
	return nil
}

// CreateUnusualStock 创建异常股票记录
func CreateUnusualStock(ctx context.Context) error {
	// 调用东方财富的接口, 来获取异常股票记录
	client := NewEastMoneyClient()
	// 1. 获取异常波动数据
	// normalStocks, err := client.GetRemoteUnusualStock(ctx)
	// if err != nil {
	// 	return err
	// }
	// if err := TryToAddUnusualStock(ctx, normalStocks); err != nil {
	// 	return err
	// }
	// 2. 获取严重异常波动数据
	specialStocks, err := client.GetRemoteSpecialUnusualStock(ctx)
	if err != nil {
		return err
	}
	if err := TryToAddUnusualStock(ctx, specialStocks); err != nil {
		return err
	}
	// 3. 获取交易所风险提示数据
	marketRiskStocks, err := client.GetRemoteMarketRisk(ctx)
	if err != nil {
		return err
	}
	if err := TryToAddUnusualStock(ctx, marketRiskStocks); err != nil {
		return err
	}
	return nil
}

// GetUnusualStockList 获取异常股票列表
func GetUnusualStockList(ctx context.Context) ([]*model.UnusualStock, error) {
	stocks, err := dal.GetUnusualStockList(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.UnusualStock, 0, len(stocks))
	for _, stock := range stocks {
		endDate := utils.FormatDate(utils.ParseDateWithRegion(stock.EndDate))
		// 如果截止日期已经超了, 就直接pass
		if utils.IsDateGreaterThan(utils.GetDateOfToday(), endDate) {
			continue
		}
		noticeDate := utils.FormatDate(utils.ParseDateWithRegion(stock.NoticeDate))
		if noticeDate == "0001-01-01" {
			noticeDate = ""
		}
		startDate := utils.FormatDate(utils.ParseDateWithRegion(stock.StartDate))
		result = append(result, &model.UnusualStock{
			Code:          stock.Code,
			Name:          stock.Name,
			Type:          model.UnusualType(stock.Type),
			StartDate:     startDate,
			EndDate:       endDate,
			NoticeDate:    noticeDate,
			UnusualType:   stock.UnusualType,
			UnusualReason: stock.UnusualReason,
		})
	}
	sort.Sort(model.UnusualStockSorter(result))
	return result, nil
}
