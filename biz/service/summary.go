package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/zhikongming/stock/biz/config"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

func GetAnalyzeReport(ctx context.Context) ([]*model.ScoreResult, error) {
	req := &model.GetIndustryTrendDataReq{
		Days: 30,
	}
	// 获取板块数据
	industryTrendList, err := GetIndustryTrendDetail(ctx, req)
	if err != nil {
		return nil, err
	}
	// 计算综合得分
	res := calculateScore(ctx, industryTrendList)
	// 发送信息通知
	message := BuildSummaryMessage(res, industryTrendList[0].PriceTrendList[0].DateString, industryTrendList[0].PriceTrendList[len(industryTrendList[0].PriceTrendList)-1].DateString)
	_ = SendLarkMessage(ctx, message)
	return res, nil
}

/*
计算各指标：
- 5日均线斜率（3日涨幅）并排名打分（10分制）。
- 10日均线斜率（3日涨幅）排名打分（10分制）。
- 判断 5>10、10>20、20>30 分别得对应分数。
- 计算20日均线斜率（5日涨幅）是否>0，得10分或0分。
- 判断收盘价是否创20日新高，得10分或0分。
- 判断成交量 > 5日均量，得10分或0分。
- 计算过去20日涨幅，得到RPS_20值，若>90得15分。
- 计算过去5日涨幅，得到RPS_5值，若>85得10分。
2. 汇总总分，得到每个板块的最终分数。
3. 筛选：输出总分≥70分的板块，或按总分排序取前15名，作为强势板块候选。若需区分异动与持续强势，可再按上述分档条件标记标签。
*/
func calculateScore(ctx context.Context, industryTrendList []*model.IndustryPriceTrend) []*model.ScoreResult {
	result := make([]*model.ScoreResult, 0, len(industryTrendList))
	resultMap := make(map[string]*model.ScoreResult)
	for _, trend := range industryTrendList {
		s := &model.ScoreResult{
			Code:  trend.IndustryCode,
			Name:  trend.IndustryName,
			Price: (trend.PriceTrendList[len(trend.PriceTrendList)-1].Price - 1) * 100,
		}
		resultMap[trend.IndustryCode] = s
		result = append(result, s)
	}
	// 根据涨跌幅区间来计算分数
	changeScoreMap := CalculateChangeScore(ctx, industryTrendList)
	for code, score := range changeScoreMap {
		s := resultMap[code]
		s.ChangeScore = score
	}
	// 5日均线斜率（3日涨幅）并排名打分（10分制）。
	codeScoreList := make([]*model.CodeScore, 0, len(industryTrendList))
	for _, trend := range industryTrendList {
		slope := CalculateSlope(5, 3, trend)
		if math.IsNaN(slope) {
			slope = 0.0
		}
		codeScore := &model.CodeScore{
			Code:  trend.IndustryCode,
			Value: slope,
		}
		codeScoreList = append(codeScoreList, codeScore)
	}
	slopeScoreMap := CalculateSlopeRankScore(codeScoreList)
	for code, score := range slopeScoreMap {
		s := resultMap[code]
		s.Slope5Score = score
	}
	// 10日均线斜率（3日涨幅）排名打分（10分制）。
	codeScoreList = make([]*model.CodeScore, 0, len(industryTrendList))
	for _, trend := range industryTrendList {
		slope := CalculateSlope(10, 3, trend)
		if math.IsNaN(slope) {
			slope = 0.0
		}
		codeScore := &model.CodeScore{
			Code:  trend.IndustryCode,
			Value: slope,
		}
		codeScoreList = append(codeScoreList, codeScore)
	}
	slopeScoreMap = CalculateSlopeRankScore(codeScoreList)
	for code, score := range slopeScoreMap {
		s := resultMap[code]
		s.Slope10Score = score
	}
	// 计算20日均线斜率（5日涨幅）是否>0，得10分或0分。
	codeScoreList = make([]*model.CodeScore, 0, len(industryTrendList))
	for _, trend := range industryTrendList {
		slope := CalculateSlope(20, 5, trend)
		if math.IsNaN(slope) {
			slope = 0.0
		}
		codeScore := &model.CodeScore{
			Code:  trend.IndustryCode,
			Value: slope,
		}
		codeScoreList = append(codeScoreList, codeScore)
	}
	slopeScoreMap = CalculateSlopeRankScore(codeScoreList)
	for code, score := range slopeScoreMap {
		s := resultMap[code]
		s.Slope20Score = score
	}
	// 判断 5>10、10>20、20>30 分别得对应分数。
	for _, trend := range industryTrendList {
		score5gt10, score10gt20, score20gt30 := ScoreMAComparisons(trend)
		s := resultMap[trend.IndustryCode]
		s.Score5gt10 = score5gt10
		s.Score10gt20 = score10gt20
		s.Score20gt30 = score20gt30
	}
	// 判断收盘价是否创20日新高，得10分或0分。
	for _, trend := range industryTrendList {
		score := ScoreNewHigh(trend)
		s := resultMap[trend.IndustryCode]
		s.NewHighScore = score
	}
	// 判断成交量 > 5日均量，得10分或0分。
	for _, trend := range industryTrendList {
		score := ScoreVolumeAboveMA5(trend)
		s := resultMap[trend.IndustryCode]
		s.VolumeScore = score
	}
	// 计算过去20日涨幅，得到RPS_20值
	codeScoreList = make([]*model.CodeScore, 0, len(industryTrendList))
	for _, trend := range industryTrendList {
		change := CalculateChange(trend, 20)
		codeScore := &model.CodeScore{
			Code:  trend.IndustryCode,
			Value: change,
		}
		codeScoreList = append(codeScoreList, codeScore)
	}
	rps20ScoreMap := CalculateChangeRankScore(codeScoreList)
	for code, score := range rps20ScoreMap {
		s := resultMap[code]
		s.RPS20Score = score
	}
	// 计算过去5日涨幅，得到RPS_5值
	codeScoreList = make([]*model.CodeScore, 0, len(industryTrendList))
	for _, trend := range industryTrendList {
		change := CalculateChange(trend, 5)
		codeScore := &model.CodeScore{
			Code:  trend.IndustryCode,
			Value: change,
		}
		codeScoreList = append(codeScoreList, codeScore)
	}
	rps5ScoreMap := CalculateChangeRankScore(codeScoreList)
	for code, score := range rps5ScoreMap {
		s := resultMap[code]
		s.RPS5Score = score
	}
	// 计算总分
	for _, s := range result {
		s.Score = utils.Float64KeepDecimal((s.ChangeScore + s.Slope5Score + s.Slope10Score + s.Slope20Score + s.Score5gt10 + s.Score10gt20 + s.Score20gt30 + s.NewHighScore + s.VolumeScore + s.RPS20Score + s.RPS5Score), 2)
	}
	// 对结果进行排序
	sort.Sort(model.ScoreResultSorter(result))
	// 只取最多15个板块
	if len(result) > 15 {
		result = result[:15]
	}
	// 计算板块内涨幅最大的股票
	industryMaxChangeStockMap := CalculateIndustryMaxChangeStock(ctx, result)
	for code, codeChange := range industryMaxChangeStockMap {
		s := resultMap[code]
		s.MaxStockCode = codeChange.Code
		s.MaxStockName = codeChange.Name
		s.MaxStockChange = codeChange.Change
	}
	// 计算板块内的第三类买点的数据
	CalculateThirdBuyPoint(ctx, result)
	return result
}

func CalculateChangeScore(ctx context.Context, trendList []*model.IndustryPriceTrend) map[string]float64 {
	resultMap := make(map[string]float64)
	MaxScore := 110.0
	maxIndustry := trendList[0]
	maxChange := maxIndustry.PriceTrendList[len(maxIndustry.PriceTrendList)-1].Price - 1
	minIndustry := trendList[len(trendList)-1]
	minChange := minIndustry.PriceTrendList[len(minIndustry.PriceTrendList)-1].Price - 1
	change := maxChange - minChange
	for _, trend := range trendList {
		score := MaxScore * (trend.PriceTrendList[len(trend.PriceTrendList)-1].Price - 1 - minChange) / change
		resultMap[trend.IndustryCode] = utils.Float64KeepDecimal(score, 2)
	}
	return resultMap
}

func CalculateThirdBuyPoint(ctx context.Context, scoreResultList []*model.ScoreResult) {
	req := &model.FilterThirdBuyCodeReq{
		Days:            30,
		ThresholdProfit: 20.0,
	}
	for _, s := range scoreResultList {
		req.IndustryName = s.Name
		thirdBuyPoint, err := FilterThirdBuyCode(ctx, req)
		if err != nil || len(thirdBuyPoint.Data) == 0 {
			continue
		}
		if len(thirdBuyPoint.Data) > 5 {
			s.ThirdBuyPoint = thirdBuyPoint.Data[:5]
		} else {
			s.ThirdBuyPoint = thirdBuyPoint.Data
		}
	}
}

func CalculateIndustryMaxChangeStock(ctx context.Context, scoreResultList []*model.ScoreResult) map[string]*model.CodeChange {
	result := make(map[string]*model.CodeChange)
	req := &model.GetIndustryTrendDataReq{
		Days: 30,
	}
	for _, s := range scoreResultList {
		req.IndustryCode = s.Code
		data, err := GetIndustryCodeDetail(ctx, req)
		if err != nil || len(data) == 0 {
			continue
		}
		result[s.Code] = &model.CodeChange{
			Code:   data[0].StockCode,
			Name:   data[0].StockName,
			Change: data[0].PriceTrendList[len(data[0].PriceTrendList)-1].Price,
		}
	}
	return result
}

func CalculateSlope(maDays int, riseDays int, trend *model.IndustryPriceTrend) float64 {
	if trend == nil || len(trend.PriceTrendList) < maDays+riseDays {
		return math.NaN()
	}
	list := trend.PriceTrendList
	n := len(list)

	// 计算最新的 maDays 日均线值
	var sumLatest float64
	for i := n - maDays; i < n; i++ {
		sumLatest += list[i].Price
	}
	maLatest := sumLatest / float64(maDays)

	// 计算 riseDays 天前的 maDays 日均线值
	// 需要定位到 riseDays 天前的数据，并取之前 maDays 个收盘价
	// 索引位置：当前最新索引为 n-1，riseDays 天前的索引为 n-1-riseDays
	idxPast := n - 1 - riseDays
	if idxPast < maDays-1 { // 检查是否有足够的数据计算过去的均线
		return math.NaN()
	}
	var sumPast float64
	for i := idxPast - maDays + 1; i <= idxPast; i++ {
		sumPast += list[i].Price
	}
	maPast := sumPast / float64(maDays)

	// 计算涨幅
	return (maLatest - maPast) / maPast
}

func ScoreFromRankPercentile(rankPercentile float64) float64 {
	if rankPercentile <= 0.2 {
		return 10.0
	}
	if rankPercentile >= 0.8 {
		return 0.0
	}
	// 线性插值：得分 = 10 * (0.8 - rankPercentile) / 0.6
	return utils.Float64KeepDecimal(10.0*(0.8-rankPercentile)/0.6, 2)
}

func ScoreFromRankRPS(rankPercentile float64) float64 {
	if rankPercentile <= 0.1 {
		return 15.0
	}
	if rankPercentile >= 0.9 {
		return 0.0
	}
	return utils.Float64KeepDecimal(15.0*(1.0-rankPercentile), 2)
}

// 得分（0~10），若数据不足或无法计算返回0
func CalculateSlopeRankScore(codeScoreList []*model.CodeScore) map[string]float64 {
	result := make(map[string]float64)
	// 对 codeScoreList 进行排序
	sort.Sort(model.CodeScoreSorter(codeScoreList))
	// 根据排序后的顺序来计算排名百分位
	for i, codeScore := range codeScoreList {
		rankPercentile := float64(i) / float64(len(codeScoreList)-1)
		score := ScoreFromRankPercentile(rankPercentile)
		result[codeScore.Code] = score
	}
	return result
}

func CalculateMA(priceList []*model.PriceTrend, days int) (float64, error) {
	if len(priceList) < days {
		return 0, errors.New("insufficient data")
	}
	sum := 0.0
	for i := len(priceList) - days; i < len(priceList); i++ {
		sum += priceList[i].Price
	}
	return sum / float64(days), nil
}

func ScoreMAComparisons(trend *model.IndustryPriceTrend) (score5gt10, score10gt20, score20gt30 float64) {
	list := trend.PriceTrendList
	// 计算各均线
	ma5, err5 := CalculateMA(list, 5)
	ma10, err10 := CalculateMA(list, 10)
	ma20, err20 := CalculateMA(list, 20)
	ma30, err30 := CalculateMA(list, 30)

	// 比较 5>10
	if err5 == nil && err10 == nil && ma5 > ma10 {
		score5gt10 = 10
	} else {
		score5gt10 = 0
	}

	// 比较 10>20
	if err10 == nil && err20 == nil && ma10 > ma20 {
		score10gt20 = 10
	} else {
		score10gt20 = 0
	}

	// 比较 20>30
	if err20 == nil && err30 == nil && ma20 > ma30 {
		score20gt30 = 10
	} else {
		score20gt30 = 0
	}
	return
}

func ScoreNewHigh(trend *model.IndustryPriceTrend) float64 {
	if trend == nil || len(trend.PriceTrendList) < 20 {
		return 0.0
	}
	list := trend.PriceTrendList
	n := len(list)

	// 取最近20个交易日（包含今天）的收盘价
	latestPrice := list[n-1].Price
	maxPrice := latestPrice
	for i := n - 20; i < n; i++ {
		if list[i].Price > maxPrice {
			maxPrice = list[i].Price
		}
	}

	// 如果最新收盘价等于最大值（即创下新高，允许并列），得10分
	if latestPrice >= maxPrice {
		return 10.0
	}
	return 0.0
}

func ScoreVolumeAboveMA5(trend *model.IndustryPriceTrend) float64 {
	if trend == nil || len(trend.PriceTrendList) < 5 {
		return 0.0
	}
	list := trend.PriceTrendList
	n := len(list)

	// 取最近5个交易日的成交量（包含今天）
	var sum int64 = 0
	for i := n - 5; i < n; i++ {
		sum += list[i].Amount
	}
	ma5 := float64(sum) / 5.0

	// 最新一天的成交量
	latestVolume := float64(list[n-1].Amount)

	// 如果最新成交量大于5日均量，得10分
	if latestVolume > ma5 {
		return 10.0
	}
	return 0.0
}

func CalculateChange(trend *model.IndustryPriceTrend, days int) float64 {
	if trend == nil || days < 1 {
		return 0.0
	}
	list := trend.PriceTrendList
	if len(list) < days+1 { // 需要至少 days+1 个数据点（今天和days天前）
		return 0.0
	}
	n := len(list)
	latestPrice := list[n-1].Price
	oldPrice := list[n-1-days].Price
	return (latestPrice - oldPrice) / oldPrice
}

func CalculateChangeRankScore(codeScoreList []*model.CodeScore) map[string]float64 {
	result := make(map[string]float64)
	// 对 codeScoreList 进行排序
	sort.Sort(model.CodeScoreSorter(codeScoreList))
	// 根据排序后的顺序来计算排名百分位
	for i, codeScore := range codeScoreList {
		rankPercentile := float64(i) / float64(len(codeScoreList)-1)
		score := ScoreFromRankRPS(rankPercentile)
		result[codeScore.Code] = score
	}
	return result
}

func BuildSummaryMessage(scoreResultList []*model.ScoreResult, dateStart, dateEnd string) *model.LarkMessage {
	// 只取前15个最强板块
	if len(scoreResultList) > 15 {
		scoreResultList = scoreResultList[:15]
	}
	industryTableElement := model.TableElement{
		Tag:       "table",
		RowHeight: "middle",
		HeaderStyle: model.HeaderStyle{
			BackgroundStyle: "none",
			Bold:            true,
			Lines:           1,
		},
		Margin:   "0px 0px 0px 0px",
		PageSize: len(scoreResultList),
		Columns: []model.Column{
			{
				DataType:        "text",
				Name:            "name",
				DisplayName:     "板块名称",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "number",
				Name:            "score",
				DisplayName:     "得分",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "price",
				DisplayName:     "涨跌幅",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "markdown",
				Name:            "operation",
				DisplayName:     "三类买点分析",
				HorizontalAlign: "left",
				Width:           "auto",
			},
		},
		Rows: make([]map[string]interface{}, 0),
	}
	for _, scoreResult := range scoreResultList {
		industryTableElement.Rows = append(industryTableElement.Rows, map[string]interface{}{
			"name":      scoreResult.Name,
			"score":     scoreResult.Score,
			"price":     fmt.Sprintf("%.2f%%", scoreResult.Price),
			"operation": fmt.Sprintf("[查看](%s/third_buy.html?sector-name=%s&days=%d)", config.GetLocalHost(), scoreResult.Name, 30),
		})
	}

	message := &model.LarkMessage{
		MsgType: "interactive",
		Card: model.LarkCard{
			Header: model.LarkHeader{
				Title: model.LarkTitle{
					Tag:     "plain_text",
					Content: "板块/股票分析总结",
				},
				Subtitle: model.LarkTitle{
					Tag:     "plain_text",
					Content: fmt.Sprintf("板块在%s - %s的表现", dateStart, dateEnd),
				},
				Template: "blue",
				Padding:  "12px 12px 12px 12px",
			},
			Schema: "2.0",
			Config: model.LarkConfig{
				UpdateMulti: true,
				Style: model.Style{
					TextSize: model.TextSize{
						NormalV2: model.NormalV2{
							Default: "medium",
							Pc:      "medium",
							Mobile:  "heading",
						},
					},
				},
			},
			Body: model.LarkBody{
				Direction:         "vertical",
				HorizontalSpacing: "8px",
				VerticalSpacing:   "8px",
				HorizontalAlign:   "left",
				VerticalAlign:     "top",
				Padding:           "12px 12px 12px 12px",
				Elements: []model.Element{
					model.MarkdownElement{
						Tag:       "markdown",
						Content:   "根据板块的综合得分, 包括均线斜率、均线位置、股价新高、成交量和RPS数据, 分析出以下板块表现出较高的潜力",
						TextAlign: "left",
						TextSize:  "normal_v2",
						Margin:    "0px 0px 0px 0px",
					},
				},
			},
		},
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, industryTableElement)

	// 分隔线
	hrElement := model.HrElement{
		Tag:    "hr",
		Margin: "0px 0px 0px 0px",
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, hrElement)

	// 股票个股
	stockMarkdownElement := model.MarkdownElement{
		Tag:       "markdown",
		Content:   "根据股票的第三类买点进行过滤, 分析出当下收盘价购买潜在利润超过25%的股票的TOP5",
		TextAlign: "left",
		TextSize:  "normal_v2",
		Margin:    "0px 0px 0px 0px",
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, stockMarkdownElement)
	stockTableElement := model.TableElement{
		Tag:       "table",
		RowHeight: "middle",
		HeaderStyle: model.HeaderStyle{
			BackgroundStyle: "none",
			Bold:            true,
			Lines:           1,
		},
		Margin:   "0px 0px 0px 0px",
		PageSize: 0,
		Columns: []model.Column{
			{
				DataType:        "text",
				Name:            "industryName",
				DisplayName:     "板块名称",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "stockName",
				DisplayName:     "股票名称",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "reupChange",
				DisplayName:     "再次上涨幅度",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "finalChange",
				DisplayName:     "潜在利润空间",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "markdown",
				Name:            "operation",
				DisplayName:     "查看",
				HorizontalAlign: "left",
				Width:           "auto",
			},
		},
		Rows: make([]map[string]interface{}, 0),
	}
	for _, scoreResult := range scoreResultList {
		for _, stockThirdBuyCodePeriodResult := range scoreResult.ThirdBuyPoint {
			stockTableElement.Rows = append(stockTableElement.Rows, map[string]interface{}{
				"industryName": scoreResult.Name,
				"stockName":    stockThirdBuyCodePeriodResult.Name,
				"reupChange":   fmt.Sprintf("%.2f%%", stockThirdBuyCodePeriodResult.ReupPeriod.Rate),
				"finalChange":  fmt.Sprintf("%.2f%%", stockThirdBuyCodePeriodResult.FinalPeriod.Rate),
				"operation":    fmt.Sprintf("[查看](https://xueqiu.com/S/%s)", stockThirdBuyCodePeriodResult.Code),
			})
		}
		stockTableElement.PageSize += len(scoreResult.ThirdBuyPoint)
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, stockTableElement)
	return message
}
