package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	ShortPeriod  = 12
	LongPeriod   = 26
	SignalPeriod = 9

	KdjRsvPeriod = 9
	KdjEmaPeriod = 3

	MaxJobNum   = 10
	MaxDBJobNum = 100
)

func GetAllCode(ctx context.Context) ([]*dal.StockCode, error) {
	codeList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}
	return codeList, nil
}

func SyncStockCode(ctx context.Context, req *model.SyncStockCodeReq) error {
	if len(req.Code) != 0 {
		return syncOneStockCode(ctx, req)
	} else {
		return syncAllStockCode(ctx, req)
	}
}

func syncOneStockCode(ctx context.Context, req *model.SyncStockCodeReq) error {
	// 判断代码是否存在
	exist, err := dal.IsStockCodeExist(ctx, req.Code)
	if err != nil {
		return err
	}
	if !exist {
		err = SyncStockBasic(ctx, req)
		if err != nil {
			return err
		}
	}

	err = SyncStockDailyPrice(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func syncAllStockCode(ctx context.Context, req *model.SyncStockCodeReq) error {
	// 获取所有股票代码
	stockCodeList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	jobs := make(chan struct{}, MaxJobNum)
	for _, stockCode := range stockCodeList {
		wg.Add(1)
		jobs <- struct{}{}
		tmpReq := model.SyncStockCodeReq{
			Code: stockCode.CompanyCode,
		}
		go SyncStockDailyPriceWrap(ctx, &tmpReq, &wg, jobs)
	}
	wg.Wait()
	return nil
}

func SyncStockBasic(ctx context.Context, req *model.SyncStockCodeReq) error {
	// 检查是否存在股票基础数据, 如果不存在就同步数据
	client := NewEastMoneyClient()
	stockBasicData, err := client.GetRemoteStockCode(ctx, req.Code)
	if err != nil {
		return err
	}
	stockCode := &dal.StockCode{
		CompanyCode:   req.Code,
		CompanyCodeHK: "",
		CompanyName:   stockBasicData.OrgShortNameCN,
		CompanyNameHK: "",
		ClassiName:    stockBasicData.ClassiName,
		BusinessType:  req.BusinessType,
		ListedDate:    utils.TimestampToDate(stockBasicData.ListedDate / 1000),
	}

	// 	获取相关股票，以确定HK代码
	stockRelationList, err := client.GetRemoteStockRelation(ctx, req.Code)
	if err != nil {
		return err
	}
	if len(stockRelationList) > 0 {
		stockCode.CompanyCodeHK = stockRelationList[0].Symbol
		stockCode.CompanyNameHK = stockRelationList[0].Name
	}

	err = dal.CreateStockCode(ctx, stockCode)
	return err
}

func SyncStockDailyPriceWrap(ctx context.Context, req *model.SyncStockCodeReq, wg *sync.WaitGroup, jobs chan struct{}) error {
	defer wg.Done()
	defer func() {
		<-jobs
	}()
	return SyncStockDailyPrice(ctx, req)
}

func GetStockPrice(ctx context.Context, code string, startTime time.Time, endTime time.Time, kLineType model.KLineType) ([]*dal.StockPrice, error) {
	// 只获取数据，不需同步数据
	client := NewEastMoneyClient()
	stockDailyData, err := client.GetRemoteStockByKLineType(ctx, code, startTime, endTime, kLineType)
	if err != nil {
		return nil, err
	}
	stockPriceList := make([]*dal.StockPrice, 0)
	for _, item := range stockDailyData.Item {
		timestampIndex := stockDailyData.GetColumnIndexByKey("timestamp")
		timestamp, _ := strconv.ParseInt(utils.ToString(item[timestampIndex]), 10, 64)
		date := utils.TimestampToDateTime(timestamp / int64(time.Microsecond))
		sp := &dal.StockPrice{
			CompanyCode: code,
			Date:        utils.ParseTime(date),
			PriceHigh:   utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("high")]), 2),
			PriceLow:    utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("low")]), 2),
			PriceOpen:   utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("open")]), 2),
			PriceClose:  utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("close")]), 2),
			Amount:      int64(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("amount")])),
			BollingUp:   0,
			BollingDown: 0,
			BollingMid:  0,
			Ma5:         0,
			Ma10:        0,
			Ma20:        0,
			Ma30:        0,
			Ma60:        0,
			MacdDif:     0,
			MacdDea:     0,
			KdjK:        0,
			KdjD:        0,
			KdjJ:        0,
		}
		stockPriceList = append(stockPriceList, sp)
	}

	CalculateMa(stockPriceList)
	CalculateBolling(stockPriceList)
	CalculateMacd(stockPriceList)
	CalculateKdj(stockPriceList)

	ret := make([]*dal.StockPrice, 0)
	for _, item := range stockPriceList {
		if item.Date.Before(startTime) || item.Date.After(endTime) {
			continue
		}
		ret = append(ret, item)
	}
	return ret, nil
}

func SyncStockDailyPrice(ctx context.Context, req *model.SyncStockCodeReq) error {
	// 检查是否存在股票基础数据, 如果不存在就同步数据
	client := NewEastMoneyClient()
	localStockDailyData, err := dal.GetLastStockPrice(ctx, req.Code)
	if err != nil {
		return err
	}
	dateTime := time.Now()
	stockDailyData, err := client.GetRemoteStockDaily(ctx, req.Code, dateTime)
	if err != nil {
		return err
	}
	stockPriceList := make([]*dal.StockPrice, 0)
	for _, item := range stockDailyData.Item {
		timestampIndex := stockDailyData.GetColumnIndexByKey("timestamp")
		timestamp, _ := strconv.ParseInt(utils.ToString(item[timestampIndex]), 10, 64)
		date := utils.TimestampToDate(timestamp / int64(time.Microsecond))
		stockPriceList = append(stockPriceList, &dal.StockPrice{
			CompanyCode: req.Code,
			Date:        utils.ParseDate(date),
			PriceHigh:   utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("high")]), 2),
			PriceLow:    utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("low")]), 2),
			PriceOpen:   utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("open")]), 2),
			PriceClose:  utils.Float64KeepDecimal(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("close")]), 2),
			Amount:      int64(utils.ToFloat64(item[stockDailyData.GetColumnIndexByKey("amount")])),
			BollingUp:   0,
			BollingDown: 0,
			BollingMid:  0,
			Ma5:         0,
			Ma10:        0,
			Ma20:        0,
			Ma30:        0,
			Ma60:        0,
			MacdDif:     0,
			MacdDea:     0,
			KdjK:        0,
			KdjD:        0,
			KdjJ:        0,
			UpdateTime:  dateTime,
		})
	}

	CalculateMa(stockPriceList)
	CalculateBolling(stockPriceList)
	CalculateMacd(stockPriceList)
	CalculateKdj(stockPriceList)
	currentTime := time.Now()
	for _, item := range stockPriceList {
		if localStockDailyData != nil && !utils.IsDateGreaterThan(utils.FormatDate(item.Date), utils.FormatDate(localStockDailyData.Date)) {
			// 检查是否需要更新
			stockPrice, err := dal.GetStockPriceByCodeAndDate(ctx, item.CompanyCode, utils.FormatDate(item.Date))
			if err != nil {
				return err
			}
			if stockPrice == nil {
				continue
			}
			if stockPrice.BollingUp == 0.0 || stockPrice.BollingDown == 0.0 || stockPrice.BollingMid == 0.0 ||
				stockPrice.Ma5 == 0.0 || stockPrice.Ma10 == 0.0 || stockPrice.Ma20 == 0.0 || stockPrice.Ma30 == 0.0 || stockPrice.Ma60 == 0.0 ||
				stockPrice.MacdDif == 0.0 || stockPrice.MacdDea == 0.0 || stockPrice.KdjK == 0.0 || stockPrice.KdjD == 0.0 || stockPrice.KdjJ == 0.0 {
				stockPrice.BollingDown = item.BollingDown
				stockPrice.BollingMid = item.BollingMid
				stockPrice.BollingUp = item.BollingUp
				stockPrice.Ma5 = item.Ma5
				stockPrice.Ma10 = item.Ma10
				stockPrice.Ma20 = item.Ma20
				stockPrice.Ma30 = item.Ma30
				stockPrice.Ma60 = item.Ma60
				stockPrice.MacdDif = item.MacdDif
				stockPrice.MacdDea = item.MacdDea
				stockPrice.KdjK = item.KdjK
				stockPrice.KdjD = item.KdjD
				stockPrice.KdjJ = item.KdjJ
				err = dal.UpdateStockPrice(ctx, stockPrice)
				if err != nil {
					return err
				}
			}
			continue
		}
		closeTime := fmt.Sprintf("%s 16:00:00", utils.FormatDate(item.Date))
		closeTimeStamp := utils.ParseTime(closeTime)
		if currentTime.After(closeTimeStamp) {
			err = dal.CreateStockPrice(ctx, item)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func CalculateMa(dailyData []*dal.StockPrice) {
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
		}
	}
}

func CalculateBolling(dailyData []*dal.StockPrice) {
	for i := 0; i < len(dailyData); i++ {
		item := dailyData[i]
		if i >= 19 {
			item.BollingMid = item.Ma20
			standardDeviation := CalculateStandardDeviation(dailyData[i-19:i+1], item.Ma20)
			item.BollingUp = utils.Float64KeepDecimal(item.Ma20+standardDeviation*2, 2)
			item.BollingDown = utils.Float64KeepDecimal(item.Ma20-standardDeviation*2, 2)
		}
	}
}

func CalculateStandardDeviation(dailyData []*dal.StockPrice, sma float64) float64 {
	sum := 0.0
	for _, item := range dailyData {
		deviation := item.PriceClose - sma
		sum += deviation * deviation
	}
	variance := sum / float64(len(dailyData))
	return math.Sqrt(variance)
}

func getPriceCloseList(dailyData []*dal.StockPrice) []float64 {
	priceCloseList := make([]float64, len(dailyData))
	for i := 0; i < len(dailyData); i++ {
		priceCloseList[i] = dailyData[i].PriceClose
	}
	return priceCloseList
}

func CalculateMacd(dailyData []*dal.StockPrice) {
	n := len(dailyData)
	priceList := getPriceCloseList(dailyData)
	ema12 := calculateEMA(priceList, ShortPeriod, true)
	ema26 := calculateEMA(priceList, LongPeriod, true)
	dif := make([]float64, n)
	for i := 0; i < n; i++ {
		dif[i] = ema12[i] - ema26[i]
	}
	dea := calculateEMA(dif, SignalPeriod, false)
	for i := 0; i < n; i++ {
		dailyData[i].MacdDif = utils.Float64KeepDecimal(dif[i], 2)
		dailyData[i].MacdDea = utils.Float64KeepDecimal(dea[i], 2)
	}
}

// 计算EMA
// data: 收盘价序列
// period: 周期（如12或26）
// initialSMA: 是否用前period日的SMA作为EMA初始值
func calculateEMA(data []float64, period int, initialSMA bool) []float64 {
	ema := make([]float64, len(data))
	if len(data) < period {
		return ema // 数据不足时返回空值
	}

	// 计算初始值（SMA或首日收盘价）
	initial := 0.0
	if initialSMA {
		for i := 0; i < period; i++ {
			initial += data[i]
		}
		initial /= float64(period)
	} else {
		initial = data[0]
	}

	// 计算EMA
	multiplier := 2.0 / (float64(period) + 1)
	ema[period-1] = initial // 初始EMA

	for i := period; i < len(data); i++ {
		ema[i] = (data[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

func CalculateKdj(dailyData []*dal.StockPrice) {
	length := len(dailyData)
	K := make([]float64, length)
	D := make([]float64, length)
	J := make([]float64, length)
	RSV := make([]float64, length)

	for i := 0; i < length; i++ {
		// 计算RSV需要至少N天的数据
		if i < KdjRsvPeriod-1 {
			RSV[i] = 0.0
			continue
		}

		// 获取最近N天的数据
		start := i - KdjRsvPeriod + 1
		if start < 0 {
			start = 0
		}
		window := dailyData[start : i+1]

		// 计算最高价和最低价
		highestHigh := window[0].PriceHigh
		lowestLow := window[0].PriceLow
		for _, d := range window {
			if d.PriceHigh > highestHigh {
				highestHigh = d.PriceHigh
			}
			if d.PriceLow < lowestLow {
				lowestLow = d.PriceLow
			}
		}

		// 计算RSV
		if math.Abs(highestHigh-lowestLow) < 1e-6 {
			RSV[i] = 0.0
		} else {
			RSV[i] = (dailyData[i].PriceClose - lowestLow) / (highestHigh - lowestLow) * 100
		}

		// 初始化K和D
		if i == KdjRsvPeriod-1 {
			K[i] = RSV[i]
			D[i] = K[i]
		} else {
			// 计算K值：前一日K的2/3 + 当日RSV的1/3
			K[i] = (2.0/3.0)*K[i-1] + (1.0/3.0)*RSV[i]
			// 计算D值：前一日D的2/3 + 当日K的1/3
			D[i] = (2.0/3.0)*D[i-1] + (1.0/3.0)*K[i]
		}

		// 计算J值
		J[i] = 3*K[i] - 2*D[i]
	}

	// 将计算结果赋值给DailyData
	for i := 0; i < length; i++ {
		dailyData[i].KdjK = utils.Float64KeepDecimal(K[i], 2)
		dailyData[i].KdjD = utils.Float64KeepDecimal(D[i], 2)
		dailyData[i].KdjJ = utils.Float64KeepDecimal(J[i], 2)
	}
}

func SyncStockIndustry(ctx context.Context, req *model.SyncStockIndustryReq) error {
	err := syncStockIndustry(ctx)
	if err != nil {
		return err
	}

	err = syncStockIndustryRelation(ctx)
	if err != nil {
		return err
	}

	// 同步个股的名称等数据
	err = syncStockCodeByIndustryRelation(ctx)
	if err != nil {
		return err
	}

	return nil
}

func syncStockIndustry(ctx context.Context) error {
	client := NewEastMoneyClient()
	// 采集板块的数据，以及板块内股票的归属数据
	remoteIndustryList, err := client.GetRemoteStockIndustry(ctx)
	if err != nil {
		return err
	}
	localIndustryList, err := dal.GetAllStockIndustry(ctx)
	if err != nil {
		return err
	}
	// 对比本地数据和远程数据，更新本地数据
	localMap := make(map[string]struct{})
	for _, localIndustry := range localIndustryList {
		localMap[localIndustry.Name] = struct{}{}
	}
	for _, remoteIndustry := range remoteIndustryList {
		if _, found := localMap[remoteIndustry.Name]; !found {
			d := &dal.StockIndustry{
				Code: remoteIndustry.Code,
				Name: remoteIndustry.Name,
			}
			if err := dal.AddStockIndustry(ctx, d); err != nil {
				return err
			}
		}
	}
	return nil
}

func syncStockIndustryRelation(ctx context.Context) error {
	localIndustryList, err := dal.GetAllStockIndustry(ctx)
	if err != nil {
		return err
	}

	dataCh := make(chan *model.WrapStockItem, len(localIndustryList))
	wg := sync.WaitGroup{}
	for _, localIndustry := range localIndustryList {
		wg.Add(1)
		go func(industry *dal.StockIndustry) {
			defer wg.Done()
			client := NewEastMoneyClient()
			remoteIndustryStockList, err := client.GetRemoteStockIndustryDetail(ctx, industry.Code)
			if err != nil {
				d := &model.WrapStockItem{
					IndustryCode: industry.Code,
					Err:          err,
				}
				dataCh <- d
				return
			}
			d := &model.WrapStockItem{
				IndustryCode: industry.Code,
				StockItem:    remoteIndustryStockList,
			}
			dataCh <- d
		}(localIndustry)
	}
	wg.Wait()
	close(dataCh)

	mp := make(map[string][]*model.StockItem)
	for d := range dataCh {
		if d.Err != nil {
			return d.Err
		}
		mp[d.IndustryCode] = d.StockItem
	}

	localStockIndustryRelationList, err := dal.GetAllStockIndustryRelation(ctx)
	if err != nil {
		return err
	}
	addList, deleteList := getDiffIndustryCode(localStockIndustryRelationList, mp)
	for _, r := range addList {
		if err := dal.AddStockIndustryRelation(ctx, r); err != nil {
			return err
		}
	}
	for _, r := range deleteList {
		if err := dal.DeleteStockIndustryRelation(ctx, r); err != nil {
			return err
		}
	}

	return nil
}

func getDiffIndustryCode(local []*dal.StockIndustryRelation, remote map[string][]*model.StockItem) ([]*dal.StockIndustryRelation, []*dal.StockIndustryRelation) {
	addList := make([]*dal.StockIndustryRelation, 0)
	// 对比本地数据和远程数据，更新本地数据
	localMap := make(map[string]struct{})
	for _, val := range local {
		tmpKey := fmt.Sprintf("%s/%s", val.IndustryCode, val.CompanyCode)
		localMap[tmpKey] = struct{}{}
	}
	for industryCode, stockList := range remote {
		for _, stock := range stockList {
			tmpKey := fmt.Sprintf("%s/%s", industryCode, stock.Code)
			if _, found := localMap[tmpKey]; !found {
				addList = append(addList, &dal.StockIndustryRelation{
					IndustryCode: industryCode,
					CompanyCode:  stock.Code,
				})
			}
		}
	}

	deleteList := make([]*dal.StockIndustryRelation, 0)
	remoteMap := make(map[string]struct{})
	for industryCode, stockList := range remote {
		for _, stock := range stockList {
			tmpKey := fmt.Sprintf("%s/%s", industryCode, stock.Code)
			remoteMap[tmpKey] = struct{}{}
		}
	}
	for _, val := range local {
		tmpKey := fmt.Sprintf("%s/%s", val.IndustryCode, val.CompanyCode)
		if _, found := remoteMap[tmpKey]; !found {
			deleteList = append(deleteList, val)
		}
	}
	return addList, deleteList
}

func syncStockCodeByIndustryRelation(ctx context.Context) error {
	localIndustryRelationList, err := dal.GetAllStockIndustryRelation(ctx)
	if err != nil {
		return err
	}
	for _, industryRelation := range localIndustryRelationList {
		exist, err := dal.IsStockCodeExist(ctx, industryRelation.CompanyCode)
		if err != nil {
			return err
		}
		if exist {
			continue
		}
		req := &model.SyncStockCodeReq{
			Code:         industryRelation.CompanyCode,
			BusinessType: 2,
		}
		err = SyncStockBasic(ctx, req)
		if err != nil {
			return err
		}
	}
	return nil
}
