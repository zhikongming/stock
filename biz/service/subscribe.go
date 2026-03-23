package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func AddSubscribeStrategyData(ctx context.Context, strategy *model.AddSubscribeStrategyReq) error {
	// To check the params.
	if strategy.StrategyType > model.StrategyTypeStockPriceChange || strategy.StrategyType < model.StrategyTypeIndustryRateChange {
		return errors.New("invalid strategy type")
	}
	if strategy.PriceChangeType > model.PriceChangeTypeLess || strategy.PriceChangeType < model.PriceChangeTypeGreater {
		return errors.New("invalid price change type")
	}
	d, _ := json.Marshal(strategy)
	data := &dal.Subscribe{
		DateTime: time.Now(),
		Strategy: string(d),
		Status:   int(dal.StatusEnabled),
	}
	return dal.CreateSubscribe(ctx, data)
}

func GetSubscribeStrategyData(ctx context.Context, strategy *model.GetSubscribeStrategyReq) ([]*model.SubscribeStrategyResult, error) {
	// 获取策略信息
	subscribeList := make([]*dal.Subscribe, 0)
	var err error
	if strategy.ID > 0 {
		subscribe, err := dal.GetSubscribeById(ctx, uint(strategy.ID))
		if err != nil {
			return nil, err
		}
		subscribeList = append(subscribeList, subscribe)
	} else {
		subscribeList, err = dal.GetAllSubscribeList(ctx)
		if err != nil {
			return nil, err
		}
	}

	// 解析策略
	subscribeStrategyResultList := make([]*model.SubscribeStrategyResult, 0)
	for _, subscribe := range subscribeList {
		var req model.AddSubscribeStrategyReq
		err := json.Unmarshal([]byte(subscribe.Strategy), &req)
		if err != nil {
			return nil, err
		}
		strategyParser, err := NewStrategyParser(ctx, &req)
		if err != nil {
			return nil, err
		}
		parseResult, err := strategyParser.Parse()
		if err != nil {
			return nil, err
		}

		subscribeStrategyResultList = append(subscribeStrategyResultList, &model.SubscribeStrategyResult{
			ID:             subscribe.ID,
			DateTime:       utils.FormatTime(subscribe.DateTime),
			StrategyType:   req.StrategyType.String(),
			Code:           parseResult.Code,
			Strategy:       parseResult.StrategyResult,
			Result:         parseResult.Result,
			StrategyDetail: strategyParser.ToSubscribeStrategyDetail(),
			LastDate:       parseResult.LastDate,
			LastResult:     subscribe.LastResult,
			Count:          subscribe.Count,
		})
	}

	return subscribeStrategyResultList, nil
}

func GetSubscribeStrategyReport(ctx context.Context) error {
	req := &model.GetSubscribeStrategyReq{}
	subscribeList, err := GetSubscribeStrategyData(ctx, req)
	if err != nil {
		return err
	}
	// 生成报告
	data := make([]*model.SubscribeStrategyResult, 0)
	for _, subscribe := range subscribeList {
		if subscribe.Result != subscribe.LastResult {
			err = dal.UpdateSubscribeResultAndCount(ctx, subscribe.ID, subscribe.Result, 1)
			if err != nil {
				return err
			}
			subscribe.Count = 1
		} else {
			err = dal.UpdateSubscribeResultAndCount(ctx, subscribe.ID, subscribe.Result, subscribe.Count+1)
			if err != nil {
				return err
			}
			subscribe.Count++
		}
		if !subscribe.Result {
			continue
		}
		data = append(data, subscribe)
	}
	// 发送信息通知
	message := BuildSubscribeStrategyReportMessage(data)
	_ = SendLarkMessage(ctx, message)
	return nil
}

func DeleteSubscribeStrategyData(ctx context.Context, strategy *model.DeleteSubscribeStrategyReq) error {
	// To check the params.
	if strategy.ID <= 0 {
		return errors.New("id must be greater than 0")
	}
	return dal.DeleteSubscribeById(ctx, uint(strategy.ID))
}
