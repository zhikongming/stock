package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

// 分析均线和布林线
func AnalyzeMa(ctx context.Context, req model.AnalyzeStockCodeReq) (*model.AnalyzeStockCodeResp, error) {
	if req.Date == "" {
		req.Date = utils.FormatDate(time.Now())
	}
	stockPriceList, err := dal.GetStockPriceByDate(ctx, req.Code, "", req.Date, 20)
	if err != nil {
		return nil, err
	}
	if len(stockPriceList) == 0 {
		return nil, fmt.Errorf("no stock price data, please sync first")
	}
	stockPriceList = utils.ListSwap(stockPriceList)

	return analyzeMaPoint(ctx, stockPriceList)
}

func analyzeMaPoint(ctx context.Context, stockPriceList []*dal.StockPrice) (*model.AnalyzeStockCodeResp, error) {
	resp := model.AnalyzeStockCodeResp{
		SuggestOperation: model.StockSuggestOperationNone,
	}
	length := len(stockPriceList)
	lastStockPrice := stockPriceList[length-1]
	// 1. 分析均线的走势
	lastSortedPriceList := []*model.MaOrderData{
		{
			MaType:  model.StockMaType5,
			MaPrice: lastStockPrice.Ma5,
		},
		{
			MaType:  model.StockMaType10,
			MaPrice: lastStockPrice.Ma10,
		},
		{
			MaType:  model.StockMaType20,
			MaPrice: lastStockPrice.Ma20,
		},
		{
			MaType:  model.StockMaType30,
			MaPrice: lastStockPrice.Ma30,
		},
		{
			MaType:  model.StockMaType60,
			MaPrice: lastStockPrice.Ma60,
		},
	}
	sort.Sort(model.SortMaOrderData(lastSortedPriceList))
	analyzeDownTrendResult, err := AnalyzeMaDownTrend(ctx, stockPriceList, lastSortedPriceList)
	if err != nil {
		return nil, err
	}
	if analyzeDownTrendResult.SuggestOperation != model.StockSuggestOperationNone {
		return analyzeDownTrendResult, nil
	}
	// 分析买点
	analyzeUpTrendResult, err := AnalyzeMaUpTrend(ctx, stockPriceList, lastSortedPriceList)
	if err != nil {
		return nil, err
	}
	if analyzeUpTrendResult.SuggestOperation != model.StockSuggestOperationNone {
		return analyzeUpTrendResult, nil
	}

	return &resp, nil
}

// 如果五日线，十日线向下，且五日线小于十日线，则认定为下跌趋势
func AnalyzeMaDownTrend(ctx context.Context, stockPriceList []*dal.StockPrice, lastSortedPriceList []*model.MaOrderData) (*model.AnalyzeStockCodeResp, error) {
	resp := model.AnalyzeStockCodeResp{
		SuggestOperation: model.StockSuggestOperationNone,
	}
	// 根据macd分析买点
	length := len(stockPriceList)

	// 分析五日均线的趋势
	ma5List := make([]float64, 0, length)
	for _, item := range stockPriceList {
		ma5List = append(ma5List, item.Ma5)
	}
	lastMinPriceIndex := FindLastMinPriceIndex(ma5List)
	if lastMinPriceIndex != length-1 {
		// 非下跌趋势，因此不予以分析
		return &resp, nil
	}
	// 分析十日均线的趋势
	ma10List := make([]float64, 0, length)
	for _, item := range stockPriceList {
		ma10List = append(ma10List, item.Ma10)
	}
	lastMinPriceIndex = FindLastMinPriceIndex(ma10List)
	if lastMinPriceIndex != length-1 {
		// 非下跌趋势，因此不予以分析
		return &resp, nil
	}

	// 五日线，十日线都处于下降趋势，且五日线小于十日线，则认定是下跌趋势
	resp.SuggestOperation = model.StockSuggestOperationSell
	resp.SuggestReason = fmt.Sprintf("ma5 and ma10 are in down trend, and %s", model.SortMaOrderData(lastSortedPriceList).GetOrder())
	resp.SuggestPriority = model.HighSuggestOperationPriority.ToInt()
	resp.MaValue = &model.MaValue{
		Ma5:  stockPriceList[length-1].Ma5,
		Ma10: stockPriceList[length-1].Ma10,
		Ma20: stockPriceList[length-1].Ma20,
		Ma30: stockPriceList[length-1].Ma30,
		Ma60: stockPriceList[length-1].Ma60,
	}
	return &resp, nil
}

func AnalyzeMaUpTrend(ctx context.Context, stockPriceList []*dal.StockPrice, lastSortedPriceList []*model.MaOrderData) (*model.AnalyzeStockCodeResp, error) {
	length := len(stockPriceList)
	resp := model.AnalyzeStockCodeResp{
		SuggestOperation: model.StockSuggestOperationNone,
		MaValue: &model.MaValue{
			Ma5:  stockPriceList[length-1].Ma5,
			Ma10: stockPriceList[length-1].Ma10,
			Ma20: stockPriceList[length-1].Ma20,
			Ma30: stockPriceList[length-1].Ma30,
			Ma60: stockPriceList[length-1].Ma60,
		},
	}

	// 分析五日均线的趋势
	ma5List := make([]float64, 0, length)
	for _, item := range stockPriceList {
		ma5List = append(ma5List, item.Ma5)
	}
	lastMinPriceIndex := FindLastMinPriceIndex(ma5List)
	if lastMinPriceIndex == length-1 {
		// 下跌趋势，因此不予以分析
		return &resp, nil
	}
	// 分析十日均线的趋势
	ma10List := make([]float64, 0, length)
	for _, item := range stockPriceList {
		ma10List = append(ma10List, item.Ma10)
	}
	lastMinPriceIndex = FindLastMinPriceIndex(ma10List)

	lastStockPrice := stockPriceList[length-1]
	if lastStockPrice.Ma5 > lastStockPrice.Ma10 {
		// 五日线大于十日线，可以认定为反转信号
		resp.SuggestOperation = model.StockSuggestOperationBuy
		resp.SuggestReason = "ma5 > ma10, 五日线反转超过十日线，因此建议买入"
		resp.SuggestPriority = model.HighSuggestOperationPriority.ToInt()
	} else {
		// 五日线小于十日线，因此还是建议观察
		resp.SuggestOperation = model.StockSuggestOperationNone
		resp.SuggestReason = "ma5 < ma10, 五日线反转，但是低于十日线，因此建议观察不做操作"
		resp.SuggestPriority = model.LowSuggestOperationPriority.ToInt()
	}
	if lastMinPriceIndex == length-1 {
		resp.SuggestReason = fmt.Sprintf("%s, m10 均线处于下降趋势", resp.SuggestReason)
	} else {
		resp.SuggestReason = fmt.Sprintf("%s, m10 均线处于上升趋势", resp.SuggestReason)
	}
	return &resp, nil
}

func AnalyzeBolling(ctx context.Context, req model.AnalyzeStockCodeReq) (*model.AnalyzeStockCodeResp, error) {
	if req.Date == "" {
		req.Date = utils.FormatDate(time.Now())
	}
	stockPriceList, err := dal.GetStockPriceByDate(ctx, req.Code, "", req.Date, 20)
	if err != nil {
		return nil, err
	}
	if len(stockPriceList) == 0 {
		return nil, fmt.Errorf("no stock price data, please sync first")
	}
	stockPriceList = utils.ListSwap(stockPriceList)

	lastStockPrice := stockPriceList[len(stockPriceList)-1]
	// 布林线只有三条线，主要是看当前所在的区间即可
	bollingValue := &model.BollingValue{
		LastBollingUp:   lastStockPrice.BollingUp,
		LastBollingDown: lastStockPrice.BollingDown,
		LastBollingMid:  lastStockPrice.BollingMid,
		LastPrice:       lastStockPrice.PriceClose,
	}
	if utils.IsClosedToHigh(lastStockPrice.PriceClose, lastStockPrice.BollingUp, lastStockPrice.BollingMid, model.BollingCmpPercent) {
		// 靠近上轨
		bollingValue.ClosedPosition = model.BollingPositionUp
		return &model.AnalyzeStockCodeResp{
			SuggestOperation: model.StockSuggestOperationHold,
			SuggestReason:    "当前价格靠近上轨，建议持有",
			SuggestPriority:  model.HighSuggestOperationPriority.ToInt(),
			BollingValue:     bollingValue,
		}, nil
	} else if utils.IsClosedToHigh(lastStockPrice.PriceClose, lastStockPrice.BollingMid, lastStockPrice.BollingDown, 1-model.BollingCmpPercent) {
		// 靠近中轨
		bollingValue.ClosedPosition = model.BollingPositionMid
		// 收盘价看由上往下穿过中轨线
		length := len(stockPriceList)
		closeList := make([]float64, 0, length)
		for _, item := range stockPriceList {
			closeList = append(closeList, item.PriceClose)
		}
		lastMinPriceIndex := FindLastMinPriceIndex(closeList)
		if lastMinPriceIndex == length-1 {
			return &model.AnalyzeStockCodeResp{
				SuggestOperation: model.StockSuggestOperationSell,
				SuggestReason:    "当前价格由上往下穿过中轨线，建议卖出或者持有观望",
				SuggestPriority:  model.LowSuggestOperationPriority.ToInt(),
				BollingValue:     bollingValue,
			}, nil
		}
		// 五日均线看由下往上穿过中轨线
		return &model.AnalyzeStockCodeResp{
			SuggestOperation: model.StockSuggestOperationBuy,
			SuggestReason:    "当前价格由下往上穿过中轨线，建议买入或者持有观望",
			SuggestPriority:  model.LowSuggestOperationPriority.ToInt(),
			BollingValue:     bollingValue,
		}, nil
	} else {
		// 靠近下轨
		bollingValue.ClosedPosition = model.BollingPositionDown
		return &model.AnalyzeStockCodeResp{
			SuggestOperation: model.StockSuggestOperationSell,
			SuggestReason:    "当前价格靠近下轨，建议卖出",
			SuggestPriority:  model.HighSuggestOperationPriority.ToInt(),
			BollingValue:     bollingValue,
		}, nil
	}
}

func AnalyzeMacd(ctx context.Context, req model.AnalyzeStockCodeReq) (*model.AnalyzeStockCodeResp, error) {
	resp := model.AnalyzeStockCodeResp{
		SuggestOperation: model.StockSuggestOperationNone,
	}
	// 根据macd分析买点
	limit := 50
	stockPriceList, err := dal.GetStockPriceByDate(ctx, req.Code, "", req.Date, limit)
	if err != nil {
		return nil, err
	}
	if len(stockPriceList) == 0 {
		return nil, fmt.Errorf("no stock price data, please sync first")
	}
	stockPriceList = utils.ListSwap(stockPriceList)
	// 分析macd

	buyPointResult := AnalyzeMacdBuyPoint(stockPriceList)
	if buyPointResult.IsBuyPoint {
		resp.SuggestOperation = model.StockSuggestOperationBuy
		resp.SuggestReason = buyPointResult.Reason
		resp.SuggestRange = fmt.Sprintf("%s - %s",
			utils.FormatDate(stockPriceList[len(stockPriceList)-buyPointResult.Length].Date),
			utils.FormatDate(stockPriceList[len(stockPriceList)-1].Date))
		resp.SuggestPriority = buyPointResult.Priority.ToInt()
		resp.MacdValue = &model.MacdValue{
			LastDif: stockPriceList[len(stockPriceList)-1].MacdDif,
			LastDea: stockPriceList[len(stockPriceList)-1].MacdDea,
			Length:  buyPointResult.Length,
		}
	}
	return &resp, nil
}

func AnalyzeMacdBuyPoint(stockPriceList []*dal.StockPrice) *model.MacdAnalyzeResult {
	// 当前数据是倒序的，所以从后往前分析
	// 1. 判断dif线的趋势，是否向上
	length := len(stockPriceList)
	deaList := make([]float64, length)
	for i, item := range stockPriceList {
		deaList[i] = item.MacdDea
	}
	lastMinPriceIndex := FindLastMinPriceIndex(deaList)
	if lastMinPriceIndex == -1 {
		return &model.MacdAnalyzeResult{
			IsBuyPoint: false,
			Length:     0,
			Reason:     "数据缺失，不足以分析macd指标",
		}
	}

	idx := lastMinPriceIndex
	for idx < length {
		if stockPriceList[idx].MacdDif >= stockPriceList[idx].MacdDea {
			break
		}
		idx++
	}
	if idx < length {
		lastMinPriceIndex = idx
	}

	if lastMinPriceIndex == length-1 {
		// 非上升序列，判断macd线指标是否为红色
		if stockPriceList[length-1].MacdDif > stockPriceList[length-1].MacdDea {
			return &model.MacdAnalyzeResult{
				IsBuyPoint: true,
				Length:     1,
				Reason:     "macd柱刚刚转红，可以适当关注买入点",
				Priority:   model.LowSuggestOperationPriority,
			}
		}
		return &model.MacdAnalyzeResult{
			IsBuyPoint: false,
			Length:     0,
			Reason:     "macd线并不是向上，非买点",
		}
	}
	// 2. 判断是否刚刚到转折点
	if lastMinPriceIndex >= length-3 {
		for idx := length - 1; idx >= lastMinPriceIndex; idx-- {
			if stockPriceList[idx].MacdDif > stockPriceList[idx].MacdDea {
				return &model.MacdAnalyzeResult{
					IsBuyPoint: true,
					Length:     length - lastMinPriceIndex - 1,
					Reason:     "macd柱刚刚转红，可以适当关注买入点",
					Priority:   model.LowSuggestOperationPriority,
				}
			}
		}
	}
	// 3. 判断dif线是否大于0
	reason := ""
	priority := model.LowSuggestOperationPriority
	if stockPriceList[length-1].MacdDif < 0 {
		if stockPriceList[length-1].MacdDif < -0.2 {
			reason = "macd线小于0，是较好的买点, 但是还未达到0轴"
		} else {
			reason = "macd线小于0，是较好的买点, 且即将过0轴"
		}

		priority = model.HighSuggestOperationPriority
	} else {
		if stockPriceList[length-1].MacdDif < 0.2 {
			reason = "macd线大于0，是较好的买点，且刚穿过0轴"
		} else {
			reason = "macd线大于0，是可以考虑的买点，, 且已经穿过了0轴"
		}
		priority = model.LowSuggestOperationPriority
	}

	return &model.MacdAnalyzeResult{
		IsBuyPoint: true,
		Length:     length - lastMinPriceIndex - 1,
		Reason:     reason,
		Priority:   priority,
	}
}

func AnalyzeKdj(ctx context.Context, req model.AnalyzeStockCodeReq) (*model.AnalyzeStockCodeResp, error) {
	// 根据kdj分析买点
	limit := 10
	stockPriceList, err := dal.GetStockPriceByDate(ctx, req.Code, "", req.Date, limit)
	if err != nil {
		return nil, err
	}
	if len(stockPriceList) <= 2 {
		return nil, fmt.Errorf("no stock price data, please sync first")
	}
	stockPriceList = utils.ListSwap(stockPriceList)

	return analyzeKdjPoint(ctx, stockPriceList)
}

func analyzeKdjPoint(ctx context.Context, stockPriceList []*dal.StockPrice) (*model.AnalyzeStockCodeResp, error) {
	// 分析kdj的趋势
	length := len(stockPriceList)
	prev := stockPriceList[length-2]
	current := stockPriceList[length-1]
	resp := model.AnalyzeStockCodeResp{
		SuggestOperation: model.StockSuggestOperationNone,
		KdjValue: &model.KdjValue{
			LastKdjK: current.KdjK,
			LastKdjD: current.KdjD,
			LastKdjJ: current.KdjJ,
		},
		SuggestPriority: model.LowSuggestOperationPriority.ToInt(),
	}

	// 检测金叉（买入信号）
	if prev.KdjK < prev.KdjD && current.KdjK > current.KdjD {
		if current.KdjK < model.KdjOversold || current.KdjD < model.KdjOversold {
			resp.SuggestOperation = model.StockSuggestOperationBuy
			resp.SuggestReason = "kdj金叉，建议买入"
		}
	}

	// 检测死叉（卖出信号）
	if prev.KdjK > prev.KdjD && current.KdjK < current.KdjD {
		if current.KdjK > model.KdjOverbought || current.KdjD > model.KdjOverbought {
			resp.SuggestOperation = model.StockSuggestOperationSell
			resp.SuggestReason = "kdj死叉，建议卖出"
		}
	}

	return &resp, nil
}

func FindLastMinPriceIndex(arr []float64) int {
	n := len(arr)
	if n == 0 {
		return -1
	}
	end := n - 1
	start := end // 初始化为最后一个元素的位置

	// 从end-1开始向前遍历，找到第一个（最远的）i，使得arr[i] < arr[end]
	for i := end - 1; i >= 0; i-- {
		if arr[i] <= arr[i+1] || arr[i]-arr[i+1] <= 0.01 {
			start = i // 继续扩展区间
		} else {
			break // 不再递增，停止遍历
		}
	}

	// 找到这个区间的最小点
	minVal := arr[start]
	for i := start + 1; i <= end; i++ {
		if arr[i] < minVal {
			minVal = arr[i]
			start = i
		}
	}

	return start
}
