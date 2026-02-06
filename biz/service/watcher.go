package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func AddWatcher(ctx context.Context, req *model.AddWatcherReq) error {
	stockCodeList := utils.ListStringIgnoreEmpty(req.StockCodeList)
	if len(stockCodeList) == 0 {
		return errors.New("stock_code_list is empty")
	}

	// 检查输入的股票数据
	stockCodeList = make([]string, 0)
	for _, code := range req.StockCodeList {
		// 根据格式检查股票代码或板块代码
		if utils.IsStockNumber(code) {
			stock := utils.GetFullStockCodeOfNumber(code)
			exist, err := dal.IsStockCodeExist(ctx, stock)
			if err != nil {
				return err
			}
			if !exist {
				return fmt.Errorf("stock code not found: %s", code)
			}
			stockCodeList = append(stockCodeList, stock)
		} else if utils.IsStockCodeWithPrefix(code) {
			stock := strings.ToUpper(code)
			exist, err := dal.IsStockCodeExist(ctx, stock)
			if err != nil {
				return err
			}
			if !exist {
				return fmt.Errorf("stock code not found: %s", code)
			}
			stockCodeList = append(stockCodeList, stock)
		} else if utils.IsIndustryCode(code) {
			industry := strings.ToUpper(code)
			industryIns, err := dal.GetStockIndustry(ctx, industry)
			if err != nil {
				return err
			}
			if industryIns != nil {
				stockCodeList = append(stockCodeList, industryIns.Code)
				continue
			}
			return fmt.Errorf("industry code not found: %s", code)
		} else {
			// 输入的是中文, 来匹配股票名称或者板块名称
			industry, err := dal.GetStockIndustryByName(ctx, code)
			if err != nil {
				return err
			}
			if industry != nil {
				stockCodeList = append(stockCodeList, industry.Code)
				continue
			}

			stock, err := dal.GetStockCodeByName(ctx, code)
			if err != nil {
				return err
			}
			if stock != nil {
				stockCodeList = append(stockCodeList, stock.CompanyCode)
				continue
			}

			return fmt.Errorf("stock or industry not found: %s", code)
		}
	}

	stockCodeList = utils.Uniq(stockCodeList)
	watcher := &dal.Watcher{
		Name:       req.Name,
		Stocks:     strings.Join(stockCodeList, ","),
		StockType:  dal.StockTypeEastmoney,
		UpdateTime: time.Now(),
		Status:     dal.StatusEnabled,
	}
	return dal.CreateWatcher(ctx, watcher)
}

func GetWatchers(ctx context.Context, req *model.GetWatchersReq) ([]*model.Watcher, error) {
	ret := make([]*model.Watcher, 0)
	if req.ID != 0 {
		watcher, err := dal.GetWatcher(ctx, uint(req.ID))
		if err != nil {
			return nil, err
		}
		if watcher == nil {
			return ret, nil
		}
		d, err := ToModel(ctx, watcher)
		if err != nil {
			return nil, err
		}
		ret = append(ret, d)
	} else {
		watcherList, err := dal.GetWatchers(ctx)
		if err != nil {
			return nil, err
		}
		for _, watcher := range watcherList {
			d, err := ToModel(ctx, watcher)
			if err != nil {
				return nil, err
			}
			ret = append(ret, d)
		}
	}
	// 对股票列表进行排序，行业在股票之前
	for _, watcher := range ret {
		sort.Sort(model.MultiCodeInfoSorter(watcher.Stocks))
	}
	return ret, nil
}

func ToModel(ctx context.Context, watch *dal.Watcher) (*model.Watcher, error) {
	stockCodeList := strings.Split(watch.Stocks, ",")
	codeNameList := make([]*model.MultiCodeInfo, 0)
	for _, code := range stockCodeList {
		if utils.IsIndustryCode(code) {
			industry, err := dal.GetStockIndustry(ctx, code)
			if err != nil {
				return nil, err
			}
			codeNameList = append(codeNameList, &model.MultiCodeInfo{
				Code: industry.Code,
				Name: industry.Name,
				Type: model.CodeTypeIndustry,
			})
		} else {
			stock, err := dal.GetStockCodeByCode(ctx, code)
			if err != nil {
				return nil, err
			}
			codeNameList = append(codeNameList, &model.MultiCodeInfo{
				Code: stock.CompanyCode,
				Name: stock.CompanyName,
				Type: model.CodeTypeStock,
			})
		}
	}
	return &model.Watcher{
		ID:         watch.ID,
		Name:       watch.Name,
		Stocks:     codeNameList,
		StockType:  watch.StockType,
		UpdateTime: watch.UpdateTime,
	}, nil
}

func DeleteWatcher(ctx context.Context, req *model.DeleteWatcherReq) error {
	return dal.DeleteWatcher(ctx, uint(req.ID))
}
