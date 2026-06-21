package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func TryToAddUnusualPredict(ctx context.Context, predicts []*model.UnusualPredict) error {
	for _, item := range predicts {
		if strings.Contains(item.Name, "退市") || strings.Contains(item.Name, "ST") {
			continue
		}
		record, err := dal.GetUnusualPredictByDateTypeRule(ctx, item.Code, item.Date, int(item.PredictType), item.RuleType)
		if err != nil {
			return err
		}
		if record != nil {
			continue
		}
		if err := dal.CreateUnusualPredict(ctx, &dal.UnusualPredict{
			Date:          item.Date,
			Code:          item.Code,
			Name:          item.Name,
			PredictType:   int(item.PredictType),
			ChangeRate:    item.ChangeRate,
			DeviationDay:  item.DeviationDay,
			DeviationRate: item.DeviationRate,
			PredictRate:   item.PredictRate,
			RuleType:      int(item.RuleType),
		}); err != nil {
			return err
		}
	}
	return nil
}

// CreateUnusualPredict 创建异动预测数据
func CreateUnusualPredict(ctx context.Context) error {
	// 调用东方财富的接口, 来获取异动预测数据
	client := NewEastMoneyClient()
	predictCodes, err := client.GetRemoteUnusualPredict(ctx)
	if err != nil {
		return err
	}
	if err := TryToAddUnusualPredict(ctx, predictCodes); err != nil {
		return err
	}
	return nil
}

// GetUnusualPredictList 获取异动预测列表
func GetUnusualPredictList(ctx context.Context) ([]*model.MergedUnusualPredict, error) {
	predict, err := dal.GetLastUnusualPredict(ctx)
	if err != nil {
		return nil, err
	}
	if predict == nil {
		return nil, nil
	}
	predicts, err := dal.GetUnusualPredictByDate(ctx, predict.Date)
	if err != nil {
		return nil, err
	}

	// 合并同一种rule的数据
	mergedPredictMap := make(map[string][]*dal.UnusualPredict)
	for _, predict := range predicts {
		key := fmt.Sprintf("%s_%d", predict.Code, predict.RuleType)
		if _, ok := mergedPredictMap[key]; !ok {
			mergedPredictMap[key] = []*dal.UnusualPredict{}
		}
		mergedPredictMap[key] = append(mergedPredictMap[key], predict)
	}

	result := make([]*model.MergedUnusualPredict, 0)
	for _, predicts := range mergedPredictMap {
		predict := predicts[0]
		item := &model.MergedUnusualPredict{
			Date:          utils.FormatDate(utils.ParseDateWithRegion(predict.Date)),
			Code:          predict.Code,
			Name:          predict.Name,
			ChangeRate:    predict.ChangeRate,
			DeviationDay:  predict.DeviationDay,
			DeviationRate: predict.DeviationRate,
			RuleType:      predict.RuleType,
			Rule:          model.GetUnusualPredictRuleDesc(predict.RuleType),
		}
		for _, predict := range predicts {
			predictType := model.PredictType(predict.PredictType)
			if predictType == model.PredictTypeNextDayUnusual {
				item.NextDayPredict = &model.PredictData{
					PredictType: predictType,
					PredictRate: predict.PredictRate,
				}
			} else {
				item.TodayPredict = &model.PredictData{
					PredictType: predictType,
					PredictRate: predict.PredictRate,
				}
			}
		}
		result = append(result, item)
	}

	// 按名称来排序即可
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
