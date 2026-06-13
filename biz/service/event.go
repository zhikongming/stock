package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

// CreateEvent 创建事件
func CreateEvent(ctx context.Context, req *model.CreateEventReq) error {
	if req.Date == "" || req.Event == "" {
		return fmt.Errorf("date and event are required")
	}

	// 处理一下股票数据
	stocks := make([]string, 0)
	slices := strings.Split(req.Stocks, ",")
	for _, item := range slices {
		stock := strings.TrimSpace(item)
		if len(stock) == 0 {
			continue
		}
		if utils.IsStockCodeWithPrefix(stock) {
			stocks = append(stocks, stock)
		} else if utils.IsStockNumber(stock) {
			stocks = append(stocks, utils.GetFullStockCodeOfNumber(stock))
		} else {
			if len(stock) < 8 {
				return fmt.Errorf("stock code %s is invalid", stock)
			}
			stock = stock[:8]
			if !utils.IsStockCodeWithPrefix(stock) {
				return fmt.Errorf("stock code %s is invalid", stock)
			}
			stocks = append(stocks, stock)
		}
	}

	event := &dal.Event{
		Date:    req.Date,
		Event:   req.Event,
		Comment: req.Comment,
		Stocks:  strings.Join(stocks, ","),
	}

	return dal.CreateEvent(ctx, event)
}

// UpdateEvent 更新事件
func UpdateEvent(ctx context.Context, req *model.UpdateEventReq) error {
	event, err := dal.GetEvent(ctx, req.ID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	if req.Date != "" {
		event.Date = req.Date
	}
	if req.Event != "" {
		event.Event = req.Event
	}
	if req.Comment != "" {
		event.Comment = req.Comment
	}
	// 处理一下股票数据
	stocks := make([]string, 0)
	slices := strings.Split(req.Stocks, ",")
	for _, item := range slices {
		stock := strings.TrimSpace(item)
		if len(stock) == 0 {
			continue
		}
		if utils.IsStockCodeWithPrefix(stock) {
			stocks = append(stocks, stock)
		} else if utils.IsStockNumber(stock) {
			stocks = append(stocks, utils.GetFullStockCodeOfNumber(stock))
		} else {
			if len(stock) < 8 {
				return fmt.Errorf("stock code %s is invalid", stock)
			}
			stock = stock[:8]
			if !utils.IsStockCodeWithPrefix(stock) {
				return fmt.Errorf("stock code %s is invalid", stock)
			}
			stocks = append(stocks, stock)
		}
	}
	event.Stocks = strings.Join(stocks, ",")

	return dal.UpdateEvent(ctx, event)
}

// DeleteEvent 删除事件
func DeleteEvent(ctx context.Context, req *model.DeleteEventReq) error {
	event, err := dal.GetEvent(ctx, req.ID)
	if err != nil {
		return err
	}
	if event == nil {
		return fmt.Errorf("event not found")
	}

	return dal.DeleteEvent(ctx, req.ID)
}

// GetEventTimeline 获取事件时间轴
func GetEventTimeline(ctx context.Context) ([]*model.TimelineEventResp, error) {
	events, err := dal.GetEvents(ctx)
	if err != nil {
		return nil, err
	}

	// 按日期分组
	dateMap := make(map[string][]*model.EventResp)
	allStockCodeList := make([]string, 0)
	for _, event := range events {
		event.Date = utils.FormatDate(utils.ParseDateWithRegion(event.Date))
		if _, ok := dateMap[event.Date]; !ok {
			dateMap[event.Date] = make([]*model.EventResp, 0)
		}
		stockCodeList := strings.Split(event.Stocks, ",")
		allStockCodeList = append(allStockCodeList, stockCodeList...)
		eventResp := &model.EventResp{
			ID:      event.ID,
			Date:    event.Date,
			Event:   event.Event,
			Comment: event.Comment,
		}
		for _, stockCode := range stockCodeList {
			eventResp.Stocks = append(eventResp.Stocks, &model.CodeBasic{
				Code: stockCode,
				Name: "",
			})
		}
		dateMap[event.Date] = append(dateMap[event.Date], eventResp)
	}
	// 填充股票数据
	allStockBasicList, err := GetCodeBasicByCodeList(ctx, allStockCodeList)
	if err != nil {
		return nil, err
	}
	for _, eventList := range dateMap {
		for _, eventResp := range eventList {
			for _, stock := range eventResp.Stocks {
				for _, stockBasic := range allStockBasicList {
					if stock.Code == stockBasic.Code {
						stock.Name = stockBasic.Name
						break
					}
				}
			}
		}
	}

	// 转换为时间轴格式
	result := make([]*model.TimelineEventResp, 0)
	for date, eventList := range dateMap {
		result = append(result, &model.TimelineEventResp{
			Date:   date,
			Events: eventList,
		})
	}

	// 按日期排序（降序）
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Date > result[i].Date {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}
