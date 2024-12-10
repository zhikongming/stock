package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func SyncStockCode(ctx context.Context, req *model.SyncStockCodeReq) error {
	var err error
	// 判断代码是否存在
	if !dal.IsStockCodeExist(req.Code) {
		err = dal.CreateStockCode(req.Code)
		if err != nil {
			return err
		}
	}

	// err = SyncStockBaiscData(ctx, req)
	// if err != nil {
	// 	return err
	// }

	err = SyncStockDailyData(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func SyncStockBaiscData(ctx context.Context, req *model.SyncStockCodeReq) error {
	// 检查是否存在股票基础数据, 如果不存在就同步数据
	if !dal.IsStockBasicExist(req.Code) {
		err := dal.CreateStockBasic(req.Code)
		if err != nil {
			return err
		}
	}
	stockBasicData, err := dal.GetStockBasic(req.Code)
	fmt.Printf("stockBasicData: %v, err: %v\n", stockBasicData, err)

	if stockBasicData == nil || err != nil {
		stockBasicData, err := GetRemoteStockBasic(ctx, req.Code)
		if err != nil {
			return err
		}
		err = dal.SaveStockBasic(req.Code, stockBasicData)
		if err != nil {
			return err
		}
	}

	return nil
}

func SyncStockDailyData(ctx context.Context, req *model.SyncStockCodeReq) error {
	// 检查是否存在股票基础数据, 如果不存在就同步数据
	if !dal.IsStockDailyExist(req.Code) {
		err := dal.CreateStockDaily(req.Code)
		if err != nil {
			return err
		}
	}
	localStockDailyData, _ := dal.GetStockDaily(req.Code)
	localStockDailyMap := make(map[string]struct{})
	for _, item := range localStockDailyData {
		localStockDailyMap[item.Date] = struct{}{}
	}

	if localStockDailyData == nil {
		localStockDailyData = make([]*model.LocalStockDailyData, 0)
	}
	dateTime := time.Now()
	stockDailyData, err := GetRemoteStockDaily(ctx, req.Code, dateTime)
	if err != nil {
		return err
	}
	for _, item := range stockDailyData.Item {
		timestampIndex := stockDailyData.GetColumnIndexByKey("timestamp")
		timestamp, _ := strconv.ParseInt(utils.ToString(item[timestampIndex]), 10, 64)
		date := utils.TimestampToDate(timestamp / int64(time.Microsecond))
		if _, ok := localStockDailyMap[date]; ok {
			continue
		}
		localStockDailyData = append(localStockDailyData, &model.LocalStockDailyData{
			Date:          date,
			PriceHigh:     utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("high")]), 2),
			PriceLow:      utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("low")]), 2),
			PriceOpen:     utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("open")]), 2),
			PriceClose:    utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("close")]), 2),
			ChangePercent: utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("percent")]), 2),
			Amount:        int64(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("amount")])),
			BollingUp:     0,
			BollingDown:   0,
			BollingMid:    0,
			Ma5:           0,
			Ma10:          0,
			Ma20:          0,
			Ma30:          0,
			Ma60:          0,
		})
	}

	CalculateMaAndBolling(localStockDailyData)
	err = dal.SaveStockDaily(req.Code, localStockDailyData)
	if err != nil {
		return err
	}

	return nil
}

func CalculateMaAndBolling(dailyData []*model.LocalStockDailyData) {
	maSum5 := 0.0
	maSum10 := 0.0
	maSum20 := 0.0
	maSum30 := 0.0
	maSum60 := 0.0
	for i := 0; i < len(dailyData); i++ {
		item := dailyData[i]
		maSum5 += item.PriceClose
		maSum10 += item.PriceClose
		maSum20 += item.PriceClose
		maSum30 += item.PriceClose
		maSum60 += item.PriceClose
		if i >= 5 {
			maSum5 -= dailyData[i-5].PriceClose
		}
		if i >= 10 {
			maSum10 -= dailyData[i-10].PriceClose
		}
		if i >= 20 {
			maSum20 -= dailyData[i-20].PriceClose
		}
		if i >= 30 {
			maSum30 -= dailyData[i-30].PriceClose
		}
		if i >= 60 {
			maSum60 -= dailyData[i-60].PriceClose
		}
		if item.Ma60 == 0.0 && i >= 19 {
			item.Ma5 = utils.Float64KeepDecimal(maSum5/5.0, 2)
			item.Ma10 = utils.Float64KeepDecimal(maSum10/10.0, 2)
			item.Ma20 = utils.Float64KeepDecimal(maSum20/20.0, 2)
			if i >= 29 {
				item.Ma30 = utils.Float64KeepDecimal(maSum30/30.0, 2)
			}
			if i >= 59 {
				item.Ma60 = utils.Float64KeepDecimal(maSum60/60.0, 2)
			}
			item.BollingMid = item.Ma20
			standardDeviation := CalculateStandardDeviation(dailyData[i-19:i+1], item.Ma20)
			item.BollingUp = utils.Float64KeepDecimal(item.Ma20+standardDeviation*2, 2)
			item.BollingDown = utils.Float64KeepDecimal(item.Ma20-standardDeviation*2, 2)
		}
	}
}

func CalculateStandardDeviation(dailyData []*model.LocalStockDailyData, sma float64) float64 {
	sum := 0.0
	for _, item := range dailyData {
		deviation := item.PriceClose - sma
		sum += deviation * deviation
	}
	variance := sum / float64(len(dailyData))
	return math.Sqrt(variance)
}
