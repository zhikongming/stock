package service

import (
	"context"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	MinClassLen = 4
)

func AnalyzeTrendCode(ctx context.Context, req model.AnalyzeTrendCodeReq) (*model.AnalyzeTrendCodeResp, error) {
	// 获取股价等数据
	if req.EndDate == "" {
		req.EndDate = utils.FormatDate(time.Now())
	}
	var stockPriceList []*dal.StockPrice
	var err error
	switch req.KLineType {
	case model.KLineTypeDay:
		stockPriceList, err = dal.GetStockPriceByDate(ctx, req.Code, req.StartDate, req.EndDate, utils.StockPriceMaxLimit)
	case model.KLineType30Min:
		startTime := utils.ParseDate(req.StartDate)
		endTime := utils.ParseDate(req.EndDate)
		var stockPriceListTmp []*dal.StockPrice
		stockPriceListTmp, err = GetStockPrice(ctx, req.Code, startTime, endTime, model.KLineType30Min)
		stockPriceList = utils.ListSwap(stockPriceListTmp)
	}
	if err != nil {
		return nil, err
	}
	stockPriceList = utils.ListSwap(stockPriceList)
	// 分析股价趋势，根据收盘价来划分为不同的上涨/下跌趋势, 划分出不同的区间出来
	trendRangeList := calTrendRangeByPriceClose(stockPriceList)
	trendRangeList = preprocessTrendRange2(stockPriceList, trendRangeList)

	// 根据分型重新计算区间段间的趋势
	trendFractalList := calTrendRangeByClass(stockPriceList, trendRangeList)

	// 根据趋势区间，计算出中枢区间
	pivotFractalList := calPivotRange(stockPriceList, trendFractalList)

	// 根据计算的背驰点，判断一二三类买卖点
	divergencePointList := calTrendDivergence(stockPriceList, trendFractalList, pivotFractalList)

	return toAnalyzeTrendCodeResp(stockPriceList, trendFractalList, pivotFractalList, divergencePointList, req.KLineType), nil
}

func toAnalyzeTrendCodeResp(stockPriceList []*dal.StockPrice, intervalList []*model.FractalInterval, pivotList []*model.PivotInterval, divergencePointList []*model.DivergencePoint, kLineType model.KLineType) *model.AnalyzeTrendCodeResp {
	ret := &model.AnalyzeTrendCodeResp{
		TrendFractal:        make([]*model.FractalItem, 0),
		PriceData:           make([]*model.PriceItem, 0),
		PivotData:           make([]*model.PivotItem, 0),
		DivergencePointData: make([]*model.DivergencePointItem, 0),
	}

	for _, item := range intervalList {
		trendItem := &model.FractalItem{
			StartDate: utils.FormatDate(stockPriceList[item.StartIndex].Date),
			EndDate:   utils.FormatDate(stockPriceList[item.EndIndex].Date),
			Class:     item.Class,
		}
		if kLineType == model.KLineType30Min {
			trendItem.StartDate = utils.FormatShortTime(stockPriceList[item.StartIndex].Date)
			trendItem.EndDate = utils.FormatShortTime(stockPriceList[item.EndIndex].Date)
		}
		if item.Class == model.ClassTop {
			trendItem.PriceStart = stockPriceList[item.StartIndex].PriceHigh
			trendItem.PriceEnd = stockPriceList[item.EndIndex].PriceLow
		} else {
			trendItem.PriceStart = stockPriceList[item.StartIndex].PriceLow
			trendItem.PriceEnd = stockPriceList[item.EndIndex].PriceHigh
		}
		ret.TrendFractal = append(ret.TrendFractal, trendItem)
	}

	for _, item := range stockPriceList {
		priceItem := &model.PriceItem{
			Date:       utils.FormatDate(item.Date),
			Amount:     item.Amount,
			PriceHigh:  item.PriceHigh,
			PriceLow:   item.PriceLow,
			PriceOpen:  item.PriceOpen,
			PriceClose: item.PriceClose,
		}
		if kLineType == model.KLineType30Min {
			priceItem.Date = utils.FormatShortTime(item.Date)
		}
		ret.PriceData = append(ret.PriceData, priceItem)
	}

	for _, item := range pivotList {
		pivotItem := &model.PivotItem{
			StartDate: utils.FormatDate(stockPriceList[item.StartIndex].Date),
			EndDate:   utils.FormatDate(stockPriceList[item.EndIndex].Date),
			PriceHigh: item.PriceHigh,
			PriceLow:  item.PriceLow,
		}
		if kLineType == model.KLineType30Min {
			pivotItem.StartDate = utils.FormatShortTime(stockPriceList[item.StartIndex].Date)
			pivotItem.EndDate = utils.FormatShortTime(stockPriceList[item.EndIndex].Date)
		}
		ret.PivotData = append(ret.PivotData, pivotItem)
	}

	for _, item := range divergencePointList {
		divergenceItem := &model.DivergencePointItem{
			Date:      utils.FormatDate(stockPriceList[item.Index].Date),
			PointType: item.PointType.ToString(),
		}
		if kLineType == model.KLineType30Min {
			divergenceItem.Date = utils.FormatShortTime(stockPriceList[item.Index].Date)
		}
		switch item.PointType {
		case model.DivergencePointBuy1, model.DivergencePointBuy2, model.DivergencePointBuy3:
			divergenceItem.Price = stockPriceList[item.Index].PriceLow
		case model.DivergencePointSell1, model.DivergencePointSell2, model.DivergencePointSell3:
			divergenceItem.Price = stockPriceList[item.Index].PriceHigh
		}
		ret.DivergencePointData = append(ret.DivergencePointData, divergenceItem)
	}

	return ret
}

func calTrendRangeByPriceClose(stockPriceList []*dal.StockPrice) []*model.TrendRange {
	n := len(stockPriceList)
	if n < 2 {
		return nil
	}

	var intervals []*model.TrendRange
	startIndex := 0
	currentTrend := model.TrendUnkown

	// 遍历所有价格变化（从i到i+1）
	for i := 0; i <= n-1; i++ {
		var diff float64
		if i == 0 {
			diff = stockPriceList[i].PriceClose - stockPriceList[i].PriceOpen
			if diff > 0 {
				currentTrend = model.TrendUp
			} else if diff < 0 {
				currentTrend = model.TrendDown
			}
			continue
		} else {
			diff = stockPriceList[i].PriceLow - stockPriceList[i-1].PriceLow
		}

		var currentDirection model.TrendType

		// 确定当前变化方向
		switch {
		case diff > 0:
			currentDirection = model.TrendUp
		case diff < 0:
			currentDirection = model.TrendDown
		default: // diff == 0
			currentDirection = currentTrend
		}

		// 处理非持平情况（上涨或下跌）
		if currentDirection != model.TrendUnkown {
			if currentTrend == model.TrendUnkown {
				// 开始新趋势
				currentTrend = currentDirection
				startIndex = i
			} else if currentDirection != currentTrend {
				// 趋势变化：结束当前趋势，开始新趋势
				item := &model.TrendRange{
					StartIndex: startIndex,
					EndIndex:   i - 1,
					Trend:      currentTrend,
				}
				if startIndex > 0 {
					item.StartIndex = startIndex - 1
				}
				intervals = append(intervals, item)
				currentTrend = currentDirection
				startIndex = i
			}
			// 趋势相同则继续，无需操作
		} else {
			// 处理持平情况
			if currentTrend != model.TrendUnkown {
				// 结束当前趋势
				item := &model.TrendRange{
					StartIndex: startIndex,
					EndIndex:   i - 1,
					Trend:      currentTrend,
				}
				if startIndex > 0 {
					item.StartIndex = startIndex - 1
				}
				intervals = append(intervals, item)
				currentTrend = model.TrendUnkown
			}
			// 持平不开始新趋势
		}
	}

	// 处理最后一个趋势（如果存在）
	if currentTrend != model.TrendUnkown {
		item := &model.TrendRange{
			StartIndex: startIndex,
			EndIndex:   n - 1,
			Trend:      currentTrend,
		}
		if startIndex > 0 {
			item.StartIndex = startIndex - 1
		}
		intervals = append(intervals, item)
	}

	return intervals
}

func preprocessTrendRange2(stockPriceList []*dal.StockPrice, originIntervalList []*model.TrendRange) []*model.TrendRange {
	// 对每三段预处理一下， 上下上，下上下分别处理
	var pre *model.TrendRange
	var cur *model.TrendRange
	var next *model.TrendRange
	n := len(originIntervalList)

	for i := 0; i < n-1; i++ {
		item := originIntervalList[i]
		if item == nil {
			continue
		}
		if pre == nil {
			pre = item
			continue
		}
		if cur == nil {
			cur = item
		}
		next = originIntervalList[i+1]
		if cur.Trend == model.TrendUp {
			// 处理该三个分区段, 格式为下上下
			if cur.EndIndex-cur.StartIndex <= 2 && stockPriceList[pre.EndIndex].PriceLow > stockPriceList[next.EndIndex].PriceLow {
				pre.EndIndex = next.EndIndex
				cur = nil
				next = nil
				originIntervalList[i] = nil
				originIntervalList[i+1] = nil
			} else {
				pre = cur
				cur = next
				next = nil
			}
		} else {
			// 处理该三个分区段, 格式为上下上
			if cur.EndIndex-cur.StartIndex <= 2 && stockPriceList[pre.EndIndex].PriceHigh < stockPriceList[next.EndIndex].PriceHigh {
				pre.EndIndex = next.EndIndex
				cur = nil
				next = nil
				originIntervalList[i] = nil
				originIntervalList[i+1] = nil
			} else {
				pre = cur
				pre = cur
				cur = next
				next = nil
			}
		}
	}
	ret := make([]*model.TrendRange, 0)
	for _, item := range originIntervalList {
		if item == nil {
			continue
		}
		ret = append(ret, item)
	}
	return ret
}

func preprocessTrendRange(stockPriceList []*dal.StockPrice, originIntervalList []*model.TrendRange) []*model.TrendRange {
	newIntervalList, isChanged := preprocessTrendRangeOnce(stockPriceList, originIntervalList)
	if !isChanged {
		return newIntervalList
	} else {
		return preprocessTrendRange(stockPriceList, newIntervalList)
	}
}

func preprocessTrendRangeOnce(stockPriceList []*dal.StockPrice, originIntervalList []*model.TrendRange) ([]*model.TrendRange, bool) {
	isChanged := false
	// 对较短的趋势区间进行合并
	intervalLenMap := make(map[int][]int)
	for i, item := range originIntervalList {
		intervalLenMap[item.EndIndex-item.StartIndex] = append(intervalLenMap[item.EndIndex-item.StartIndex], i)
	}
	for i := 0; i <= 2; i++ {
		indexList, ok := intervalLenMap[i]
		if !ok {
			continue
		}
		for _, idx := range indexList {
			if idx > 0 && (originIntervalList[idx].EndIndex-originIntervalList[idx].StartIndex) < (originIntervalList[idx-1].EndIndex-originIntervalList[idx-1].StartIndex) {
				originIntervalList[idx].Trend = originIntervalList[idx-1].Trend
				isChanged = true
			} else if idx < len(originIntervalList)-1 && (originIntervalList[idx].EndIndex-originIntervalList[idx].StartIndex) < (originIntervalList[idx+1].EndIndex-originIntervalList[idx+1].StartIndex) {
				originIntervalList[idx].Trend = originIntervalList[idx+1].Trend
				isChanged = true
			}
		}
	}
	// 合并区间
	ret := make([]*model.TrendRange, 0)
	var lastInterval *model.TrendRange
	for _, item := range originIntervalList {
		if lastInterval == nil {
			lastInterval = item
			ret = append(ret, lastInterval)
			continue
		}
		if lastInterval.Trend == item.Trend {
			lastInterval.EndIndex = item.EndIndex
		} else {
			lastInterval = item
			ret = append(ret, lastInterval)
		}
	}
	return ret, isChanged
}

func calTrendRangeByClass(stockPriceList []*dal.StockPrice, originIntervalList []*model.TrendRange) []*model.FractalInterval {
	// 根据是否构造出来顶分型和底分型， 来判断出是否构成趋势
	// 顶分型： 价格从下往上， 底分型： 价格从上往下
	ret := make([]*model.FractalInterval, 0)
	n := len(originIntervalList)
	var lastFactalInterval *model.FractalInterval
	i := 0
	for i < n {
		if lastFactalInterval == nil {
			// 初始下跌判定为顶分型，初始上涨判定为底分型
			if originIntervalList[i].Trend == model.TrendDown {
				lastFactalInterval = &model.FractalInterval{
					StartIndex: originIntervalList[i].StartIndex,
					EndIndex:   originIntervalList[i].EndIndex,
					Class:      model.ClassTop,
				}
			} else if originIntervalList[i].Trend == model.TrendUp {
				lastFactalInterval = &model.FractalInterval{
					StartIndex: originIntervalList[i].StartIndex,
					EndIndex:   originIntervalList[i].EndIndex,
					Class:      model.ClassBottom,
				}
			}
			ret = append(ret, lastFactalInterval)
			i++
			continue
		}

		if lastFactalInterval.Class == model.ClassBottom {
			// 上一个是底分型，本次尝试构造顶分型
			if originIntervalList[i].EndIndex-originIntervalList[i].StartIndex < MinClassLen && i < n-1 {
				// 区间长度小于5且不为最后一段上涨， 不构成顶分型， 继续保持底分型
				lastFactalInterval.EndIndex = originIntervalList[i+1].EndIndex
				i += 2
			} else {
				// 构成顶分型
				lastFactalInterval = &model.FractalInterval{
					StartIndex: originIntervalList[i].StartIndex,
					EndIndex:   originIntervalList[i].EndIndex,
					Class:      model.ClassTop,
				}
				i++
				ret = append(ret, lastFactalInterval)
			}
		} else {
			// 上一个是顶分型，本次尝试构造底分型
			if originIntervalList[i].EndIndex-originIntervalList[i].StartIndex < MinClassLen && i < n-1 {
				// 区间长度小于5且不为最后一段下跌， 不构成底分型， 继续保持顶分型
				lastFactalInterval.EndIndex = originIntervalList[i+1].EndIndex
				i += 2
			} else {
				// 构成底分型
				lastFactalInterval = &model.FractalInterval{
					StartIndex: originIntervalList[i].StartIndex,
					EndIndex:   originIntervalList[i].EndIndex,
					Class:      model.ClassBottom,
				}
				i++
				ret = append(ret, lastFactalInterval)
			}
		}
	}

	adjustFractalInterval(stockPriceList, ret)

	return ret
}

func adjustFractalInterval(stockPriceList []*dal.StockPrice, intervalList []*model.FractalInterval) {
	var pre *model.FractalInterval
	var cur *model.FractalInterval
	n := len(intervalList)

	for i := 0; i < n-1; i++ {
		item := intervalList[i]
		if item == nil {
			continue
		}
		if pre == nil {
			pre = item
			continue
		}
		cur = item
		if cur.Class == model.ClassBottom {
			// 处理该三个分区段, 格式为上下上
			minIdx := cur.StartIndex
			for idx := cur.StartIndex; idx < cur.EndIndex; idx++ {
				// 找到区间内的最低点
				if stockPriceList[idx].PriceLow < stockPriceList[minIdx].PriceLow {
					minIdx = idx
				}
			}
			pre.EndIndex = minIdx
			cur.StartIndex = minIdx
			pre = cur
		} else {
			// 处理该三个分区段, 格式为下上下
			maxIdx := cur.StartIndex
			for idx := cur.StartIndex; idx < cur.EndIndex; idx++ {
				// 找到区间内的最高点
				if stockPriceList[idx].PriceHigh > stockPriceList[maxIdx].PriceHigh {
					maxIdx = idx
				}
			}
			pre.EndIndex = maxIdx
			cur.StartIndex = maxIdx
			pre = cur
		}
	}
}

func calPivotRange(stockPriceList []*dal.StockPrice, intervalList []*model.FractalInterval) []*model.PivotInterval {
	ret := make([]*model.PivotInterval, 0)
	n := len(intervalList)
	i := 0
	for i <= n-3 {
		t1 := intervalList[i]
		t2 := intervalList[i+1]
		t3 := intervalList[i+2]

		l1, h1 := getPriceLowAndHigh(stockPriceList, t1)
		l2, h2 := getPriceLowAndHigh(stockPriceList, t2)
		l3, h3 := getPriceLowAndHigh(stockPriceList, t3)

		zoneHigh := min(h1, h2, h3)
		zoneLow := max(l1, l2, l3)
		if zoneLow <= zoneHigh {
			// 创建新中枢
			zone := &model.PivotInterval{
				StartIndex: t1.StartIndex,
				PriceHigh:  zoneHigh,
				EndIndex:   t3.StartIndex,
				PriceLow:   zoneLow,
			}

			// 尝试扩展中枢
			j := i + 3
			for j < n {
				curr := intervalList[j]
				lowCurr, highCurr := getPriceLowAndHigh(stockPriceList, curr)

				// 检查当前笔是否与中枢重叠
				if lowCurr <= zone.PriceHigh && highCurr >= zone.PriceLow {
					// 更新中枢结束时间
					zone.EndIndex = curr.StartIndex
					j++
				} else {
					break
				}
			}

			// 保存完整中枢
			ret = append(ret, zone)
			// 跳过已处理的中枢笔
			i = j
		} else {
			// 无重叠则向前移动
			i++
		}
	}

	return ret
}

func getPriceLowAndHigh(stockPriceList []*dal.StockPrice, frac *model.FractalInterval) (float64, float64) {
	if frac.Class == model.ClassTop {
		// 属于下上下类型
		return stockPriceList[frac.EndIndex].PriceLow, stockPriceList[frac.StartIndex].PriceHigh
	} else {
		// 属于上下上类型
		return stockPriceList[frac.StartIndex].PriceLow, stockPriceList[frac.EndIndex].PriceHigh
	}
}

func calTrendDivergence(stockPriceList []*dal.StockPrice, intervalList []*model.FractalInterval, pivotFractalList []*model.PivotInterval) []*model.DivergencePoint {
	if len(pivotFractalList) == 0 {
		return nil
	}
	// 根据股价创新低/高来判断背驰点
	ret := make([]*model.DivergencePoint, 0)
	/*
	* 背驰：
	* 	趋势背驰：在同方向的两个连续中枢（如上涨趋势的两个上涨中枢，或下跌趋势的两个下跌中枢）之后，出现背驰段。
	*		第一中枢后出现离开段（a段）；随后形成第二中枢；第二中枢后再出现第二次离开段（b段）；对比a段与b段的力度（如MACD红绿柱面积、黄白线高度、K线斜率等）。
	*		结论：若b段力度弱于a段，则形成标准趋势背驰，预示原趋势结束
	*	盘整背驰：在一个中枢震荡内部，比较进入段与离开段的力度。
	*		第一笔下跌进入中枢（A段）；中枢震荡后，第二笔下跌离开中枢（B段）；若B段力度弱于A段（如MACD绿柱缩小、底背离），则形成盘整背驰
	 */
	// 如果有多个区间的话，则找到最低的那个区间，然后判断是否背驰即可。
	checkPivotFractal := pivotFractalList[0]
	direction := model.TrendUnkown
	if len(pivotFractalList) > 1 {
		nextPivotFractal := pivotFractalList[1]
		if nextPivotFractal.PriceLow < checkPivotFractal.PriceLow {
			// 中枢下降
			for _, item := range pivotFractalList {
				if item.PriceLow < checkPivotFractal.PriceLow {
					checkPivotFractal = item
				}
			}
			direction = model.TrendDown
		} else {
			// 中枢上升
			for _, item := range pivotFractalList {
				if item.PriceLow > checkPivotFractal.PriceLow {
					checkPivotFractal = item
				}
			}
			direction = model.TrendUp
		}
		ret = findDivergence(stockPriceList, checkPivotFractal, intervalList, direction)
	} else if len(pivotFractalList) == 1 {
		// 不确定中枢是那个，所以两个方向都尝试下
		direction = model.TrendUp
		retUp := findDivergence(stockPriceList, checkPivotFractal, intervalList, direction)
		ret = append(ret, retUp...)
		direction = model.TrendDown
		retDown := findDivergence(stockPriceList, checkPivotFractal, intervalList, direction)
		ret = append(ret, retDown...)
	}

	return ret
}

func findDivergence(stockPriceList []*dal.StockPrice, pivotFractal *model.PivotInterval, intervalList []*model.FractalInterval, direction model.TrendType) []*model.DivergencePoint {
	ret := make([]*model.DivergencePoint, 0)
	// 计算第一类买点
	firstDivergence := findFirstDivergence(stockPriceList, pivotFractal, intervalList, direction)
	if firstDivergence != nil {
		ret = append(ret, firstDivergence)
		// 计算第二类买点
		secondDivergence := findSecondDivergence(stockPriceList, pivotFractal, intervalList, direction, firstDivergence)
		if secondDivergence != nil {
			ret = append(ret, secondDivergence)
		}
	}
	return ret
}

func findFirstDivergence(stockPriceList []*dal.StockPrice, pivotFractal *model.PivotInterval, intervalList []*model.FractalInterval, direction model.TrendType) *model.DivergencePoint {
	// 该中枢区间计算，分为两部分：入与出
	// 入： 进入该中枢，一定是计算的入笔
	// 出： 有两种可能性，一种是该中枢结束，计算的是离开的一笔，另一种是该中枢内震荡，在震荡中离开该中枢，后续又给拉回到中枢内，属于盘整背驰
	switch direction {
	case model.TrendUp:
		// 上升趋势
		trendList := make([]*model.FractalInterval, 0)
		for _, item := range intervalList {
			if item.StartIndex < pivotFractal.StartIndex {
				continue
			}
			if item.Class == model.ClassBottom {
				trendList = append(trendList, item)
			}
		}
		// 根据进入该中枢的一笔，以及递减下跌的笔判断是否有背驰的点。
		maxFractalInterval := getMaxFractalInterval(stockPriceList, trendList)
		enterFractalInterval := getFractalIntervalByEndIndex(stockPriceList, pivotFractal.StartIndex, intervalList)
		if maxFractalInterval == nil || enterFractalInterval == nil || stockPriceList[maxFractalInterval.EndIndex].PriceHigh <= stockPriceList[enterFractalInterval.EndIndex].PriceHigh {
			return nil
		}
		// 判断两段笔是否构成背驰
		enterMacdAreaValue, enterMacdDifValue := CalMacdResult(stockPriceList, enterFractalInterval, direction)
		maxMacdAreaValue, maxMacdDifValue := CalMacdResult(stockPriceList, maxFractalInterval, direction)
		if maxMacdAreaValue < enterMacdAreaValue && (maxMacdDifValue < enterMacdDifValue || maxMacdDifValue-enterMacdDifValue < 0.04) {
			// 构成背驰
			return &model.DivergencePoint{
				Index:     maxFractalInterval.EndIndex,
				PointType: model.DivergencePointSell1,
			}
		}
	case model.TrendDown:
		// 下降趋势
		// 找到所有的下降趋势，并根据下降趋势，找到最大的下降子序列。
		trendList := make([]*model.FractalInterval, 0)
		for _, item := range intervalList {
			if item.StartIndex < pivotFractal.StartIndex {
				continue
			}
			if item.Class == model.ClassTop {
				trendList = append(trendList, item)
			}
		}
		// 根据进入该中枢的一笔，以及递减下跌的笔判断是否有背驰的点。
		minFractalInterval := getMinFractalInterval(stockPriceList, trendList)
		enterFractalInterval := getFractalIntervalByEndIndex(stockPriceList, pivotFractal.StartIndex, intervalList)
		if minFractalInterval == nil || enterFractalInterval == nil || stockPriceList[minFractalInterval.EndIndex].PriceLow >= stockPriceList[enterFractalInterval.EndIndex].PriceLow {
			return nil
		}
		// 判断两段笔是否构成背驰
		enterMacdAreaValue, enterMacdDifValue := CalMacdResult(stockPriceList, enterFractalInterval, direction)
		minMacdAreaValue, minMacdDifValue := CalMacdResult(stockPriceList, minFractalInterval, direction)
		if minMacdAreaValue > enterMacdAreaValue && (minMacdDifValue > enterMacdDifValue || enterMacdDifValue-minMacdDifValue < 0.04) {
			// 构成背驰
			return &model.DivergencePoint{
				Index:     minFractalInterval.EndIndex,
				PointType: model.DivergencePointBuy1,
			}
		}
	default:
		return nil
	}

	return nil
}

func findSecondDivergence(stockPriceList []*dal.StockPrice, pivotFractal *model.PivotInterval, intervalList []*model.FractalInterval, direction model.TrendType, firstDivergencePoint *model.DivergencePoint) *model.DivergencePoint {
	var firstDivergencePointInterval *model.FractalInterval
	for _, item := range intervalList {
		if item.EndIndex == firstDivergencePoint.Index {
			firstDivergencePointInterval = item
			continue
		}
		if firstDivergencePointInterval != nil {
			switch direction {
			case model.TrendUp:
				if item.Class == model.ClassBottom {
					return &model.DivergencePoint{
						Index:     item.EndIndex,
						PointType: model.DivergencePointSell2,
					}
				}
			case model.TrendDown:
				if item.Class == model.ClassTop {
					return &model.DivergencePoint{
						Index:     item.EndIndex,
						PointType: model.DivergencePointBuy2,
					}
				}
			}
		}
	}
	return nil
}

func getFractalIntervalByEndIndex(stockPriceList []*dal.StockPrice, idx int, intervalList []*model.FractalInterval) *model.FractalInterval {
	for _, item := range intervalList {
		if item.EndIndex == idx {
			return item
		}
	}
	// 如果没有找到笔的终点为idx的笔， 则找到起点为idx的一笔
	for _, item := range intervalList {
		if item.StartIndex == idx {
			return item
		}
	}
	return nil
}

func longestDescendingSegments(stockPriceList []*dal.StockPrice, trendList []*model.FractalInterval) []*model.FractalInterval {
	n := len(trendList)
	if n == 0 {
		return nil
	}

	// dp[i]: 以第i个线段结尾的最长递增子序列长度
	dp := make([]int, n)
	// prev[i]: 在最长子序列中，第i个线段的前驱索引
	prev := make([]int, n)
	maxLen := 0
	endIdx := -1

	// 初始化dp和prev数组
	for i := 0; i < n; i++ {
		dp[i] = 1
		prev[i] = -1
		for j := 0; j < i; j++ {
			// 检查结束位置是否严格递减
			if stockPriceList[trendList[j].EndIndex].PriceLow > stockPriceList[trendList[i].EndIndex].PriceLow {
				if dp[j]+1 > dp[i] {
					dp[i] = dp[j] + 1
					prev[i] = j
				}
			}
		}
		// 更新最大长度和结束索引
		if dp[i] > maxLen {
			maxLen = dp[i]
			endIdx = i
		}
	}

	// 回溯构建最长子序列
	result := make([]*model.FractalInterval, 0, maxLen)
	for endIdx != -1 {
		result = append(result, trendList[endIdx])
		endIdx = prev[endIdx]
	}

	// 反转结果序列
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func getMinFractalInterval(stockPriceList []*dal.StockPrice, intervalList []*model.FractalInterval) *model.FractalInterval {
	if len(intervalList) == 0 {
		return nil
	}
	minFractalInterval := intervalList[0]
	for _, item := range intervalList {
		if stockPriceList[item.EndIndex].PriceLow < stockPriceList[minFractalInterval.EndIndex].PriceLow {
			minFractalInterval = item
		}
	}
	return minFractalInterval
}

func getMaxFractalInterval(stockPriceList []*dal.StockPrice, intervalList []*model.FractalInterval) *model.FractalInterval {
	if len(intervalList) == 0 {
		return nil
	}
	maxFractalInterval := intervalList[0]
	for _, item := range intervalList {
		if stockPriceList[item.EndIndex].PriceLow > stockPriceList[maxFractalInterval.EndIndex].PriceLow {
			maxFractalInterval = item
		}
	}
	return maxFractalInterval
}

func CalMacdResult(stockPriceList []*dal.StockPrice, trend *model.FractalInterval, direction model.TrendType) (area float64, difValue float64) {
	switch direction {
	case model.TrendDown:
		// 下降趋势, 主要计算绿柱子的面积，以及dif的最小值
		trendArea := 0.0
		minMacdDif := 0.0
		for idx := trend.StartIndex; idx <= trend.EndIndex; idx++ {
			if stockPriceList[idx].GetMacdValue() < 0.0 {
				trendArea += stockPriceList[idx].GetMacdValue()
				minMacdDif = min(minMacdDif, stockPriceList[idx].MacdDif)
			}
		}
		return trendArea, minMacdDif
	case model.TrendUp:
		// 上升趋势， 主要计算红柱子的面积，以及dif的最大值
		trendArea := 0.0
		maxMacdDif := 0.0
		for idx := trend.StartIndex; idx <= trend.EndIndex; idx++ {
			if stockPriceList[idx].GetMacdValue() > 0.0 {
				trendArea += stockPriceList[idx].GetMacdValue()
				maxMacdDif = max(maxMacdDif, stockPriceList[idx].MacdDif)
			}
		}
		return trendArea, maxMacdDif
	default:
		return 0.0, 0.0
	}
}
