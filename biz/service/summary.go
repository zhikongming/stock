package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/zhikongming/stock/biz/config"
	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	VolumeReportDiffThreshold = 1.8
	MaxVolumeReportJobNum     = 50
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
	// 获取上一个分数并计算diff
	scoreDiff := getScoreDiff(ctx, res, industryTrendList[0].PriceTrendList[len(industryTrendList[0].PriceTrendList)-2].DateString)
	// 计算这些股票的量价关系
	calculatePriceAnalyse(ctx, res)
	// 发送信息通知
	message := BuildSummaryMessage(res, industryTrendList[0].PriceTrendList[0].DateString, industryTrendList[0].PriceTrendList[len(industryTrendList[0].PriceTrendList)-1].DateString, scoreDiff)
	_ = SendLarkMessage(ctx, message)
	setScoreCache(ctx, res, industryTrendList[0].PriceTrendList[len(industryTrendList[0].PriceTrendList)-1].DateString)
	return res, nil
}

func calculatePriceAnalyse(ctx context.Context, scoreList []*model.ScoreResult) error {
	codeList := make([]string, 0, len(scoreList))
	nameList := make([]string, 0, len(scoreList))
	for _, score := range scoreList {
		for _, thirdBuyPoint := range score.ThirdBuyPoint {
			codeList = append(codeList, thirdBuyPoint.Code)
			nameList = append(nameList, thirdBuyPoint.Name)
		}
	}
	cozeCache := GetCozeCache()
	volumePriceList, err := cozeCache.GetMultiVolumePrice(ctx, codeList, nameList)
	if err != nil {
		return err
	}
	// 解析数据
	volumePriceMap := make(map[string]string)
	for _, volumePrice := range volumePriceList {
		volumePriceMap[volumePrice.CompanyCode] = volumePrice.IsSafe
	}
	for _, score := range scoreList {
		for _, thirdBuyPoint := range score.ThirdBuyPoint {
			if _, ok := volumePriceMap[thirdBuyPoint.Code]; ok {
				thirdBuyPoint.PriceAnalyseResult = volumePriceMap[thirdBuyPoint.Code]
			} else {
				thirdBuyPoint.PriceAnalyseResult = "-"
			}
		}
	}
	return nil
}

func setScoreCache(ctx context.Context, scoreList []*model.ScoreResult, date string) {
	// 获取上一个交易日的缓存数据
	cache, err := dal.GetCacheByTypeDate(ctx, string(dal.CacheKeyScoreResult), dal.CacheTypeScoreResult, date)
	if err != nil || cache != nil {
		return
	}
	// 缓存数据
	data := make([]*model.SimpleScoreResult, 0, len(scoreList))
	for _, score := range scoreList {
		data = append(data, &model.SimpleScoreResult{
			Code:  score.Code,
			Name:  score.Name,
			Score: score.Score,
		})
	}
	scoreByte, _ := json.Marshal(data)
	cache = &dal.Cache{
		DataKey:   string(dal.CacheKeyScoreResult),
		DataType:  int8(dal.CacheTypeScoreResult),
		Date:      date,
		DataValue: string(scoreByte),
	}
	dal.CreateCache(ctx, cache)
	return
}

func getScoreDiff(ctx context.Context, scoreList []*model.ScoreResult, date string) map[string]*model.ScoreResultDiff {
	diffMap := make(map[string]*model.ScoreResultDiff)
	for _, score := range scoreList {
		diffMap[score.Code] = &model.ScoreResultDiff{
			ScoreDiff: "-",
			OrderDiff: "-",
		}
	}
	// 获取上一个交易日的缓存数据
	cache, err := dal.GetCacheByTypeDate(ctx, string(dal.CacheKeyScoreResult), dal.CacheTypeScoreResult, date)
	if err != nil || cache == nil {
		return diffMap
	}
	// 解析缓存数据
	var prevScoreList []*model.SimpleScoreResult
	err = json.Unmarshal([]byte(cache.DataValue), &prevScoreList)
	if err != nil {
		return diffMap
	}
	// 计算diff
	for idx, score := range scoreList {
		match := false
		for preIdx, preScore := range prevScoreList {
			if preScore.Code == score.Code {
				diffMap[score.Code].ScoreDiff = getScoreDiffMessage(preScore.Score, score.Score)
				diffMap[score.Code].OrderDiff = getOrderDiffMessage(preIdx, idx)
				match = true
				break
			}
		}
		if !match {
			diffMap[score.Code].ScoreDiff = "-"
			diffMap[score.Code].OrderDiff = "新进"
		}
	}
	return diffMap
}

func getScoreDiffMessage(originScore float64, newScore float64) string {
	if originScore == newScore {
		return "-"
	}
	if newScore > originScore {
		return fmt.Sprintf("+%.2f", newScore-originScore)
	}
	return fmt.Sprintf("-%.2f", originScore-newScore)
}

func getOrderDiffMessage(originOrder int, newOrder int) string {
	if originOrder == newOrder {
		return "-"
	}
	if originOrder > newOrder {
		return fmt.Sprintf("上升 %d 位", originOrder-newOrder)
	}
	return fmt.Sprintf("下降 %d 位", newOrder-originOrder)
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

func BuildSummaryMessage(scoreResultList []*model.ScoreResult, dateStart, dateEnd string, scoreDiff map[string]*model.ScoreResultDiff) *model.LarkMessage {
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
				DataType:        "text",
				Name:            "score_change",
				DisplayName:     "得分变化",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "order_change",
				DisplayName:     "排名变化",
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
			"name":         scoreResult.Name,
			"score":        scoreResult.Score,
			"price":        fmt.Sprintf("%.2f%%", scoreResult.Price),
			"score_change": scoreDiff[scoreResult.Code].ScoreDiff,
			"order_change": scoreDiff[scoreResult.Code].OrderDiff,
			"operation":    fmt.Sprintf("[查看](%s/third_buy.html?sector-name=%s&days=%d)", config.GetLocalHost(), scoreResult.Name, 30),
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
				DataType:        "text",
				Name:            "priceAnalyseResult",
				DisplayName:     "量价关系分析",
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
				"industryName":       scoreResult.Name,
				"stockName":          stockThirdBuyCodePeriodResult.Name,
				"reupChange":         fmt.Sprintf("%.2f%%", stockThirdBuyCodePeriodResult.ReupPeriod.Rate),
				"finalChange":        fmt.Sprintf("%.2f%%", stockThirdBuyCodePeriodResult.FinalPeriod.Rate),
				"priceAnalyseResult": stockThirdBuyCodePeriodResult.PriceAnalyseResult,
				"operation":          fmt.Sprintf("[查看](https://xueqiu.com/S/%s)", stockThirdBuyCodePeriodResult.Code),
			})
		}
		stockTableElement.PageSize += len(scoreResult.ThirdBuyPoint)
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, stockTableElement)
	return message
}

func GetPriceAnalyseReport(ctx context.Context) (*model.PriceAnalyseReport, error) {
	// 获取缓存中的最后几个分析结果
	cozeCache := GetCozeCache()
	cacheList, err := cozeCache.GetLastNMultiVolumePrice(ctx, AnalyzeVolumePriceReportLimit)
	if err != nil {
		return nil, err
	}
	if len(cacheList) == 0 {
		return nil, fmt.Errorf("no analyze result found")
	}
	// 获取需要分析的股票列表
	stockCodeList, err := dal.GetStockCodeByParsedPrice(ctx, true)
	if err != nil {
		return nil, err
	}
	codeMap := make(map[string][]string)
	for _, stockCode := range stockCodeList {
		codeMap[stockCode.CompanyName] = make([]string, 0)
	}
	// 解析数据
	for _, cache := range cacheList {
		// 解析缓存数据
		var results []*model.MultiVolumePrice
		err = json.Unmarshal([]byte(cache.DataValue), &results)
		if err != nil {
			return nil, err
		}
		for _, item := range results {
			if _, ok := codeMap[item.CompanyName]; !ok {
				continue
			}
			if item.IsSafe == IsSafeDirtyStatus {
				var result model.MultiVolumePrice
				sanitized := strings.ReplaceAll(item.AnalysisResult, "\n", "\\n")
				err = json.Unmarshal([]byte(sanitized), &result)
				if err == nil {
					item.IsSafe = result.IsSafe
					item.AnalysisResult = result.AnalysisResult
				}
			}
			codeMap[item.CompanyName] = append(codeMap[item.CompanyName], item.IsSafe)
		}
	}
	result := &model.PriceAnalyseReport{
		EndDate: cacheList[0].Date,
		Items:   make([]*model.PriceAnalyseReportItem, 0),
	}
	// 统计每只股票连续状态的天数
	for companyName, isSafeList := range codeMap {
		isSafeResult := "-"
		if len(isSafeList) > 0 {
			isSafeResult = isSafeList[0]
		}
		count := 1
		for idx := 1; idx < len(isSafeList); idx++ {
			if isSafeList[idx] == isSafeResult {
				count++
			} else {
				break
			}
		}
		result.Items = append(result.Items, &model.PriceAnalyseReportItem{
			Name:   companyName,
			IsSafe: isSafeResult,
			Count:  count,
		})
	}
	// 发送信息通知
	message := BuildPriceAnalyseReportMessage(result)
	_ = SendLarkMessage(ctx, message)
	return result, nil
}

func GetVolumeReport(ctx context.Context) ([]*model.VolumeReportItem, error) {
	// 根据最近的两个交易日的数据, 来计算是否出现成交量大幅增加
	stockList, err := dal.GetAllStockCode(ctx)
	if err != nil {
		return nil, err
	}
	// 需要使用并发来计算, 以减少耗时
	jobList := make([]func() (interface{}, error), 0)
	for _, stockCode := range stockList {
		jobList = append(jobList, func(stockCode *dal.StockCode) func() (interface{}, error) {
			return func() (interface{}, error) {
				stockPriceList, err := dal.GetLastNStockPrice(ctx, stockCode.CompanyCode, "", 2)
				if err != nil {
					return nil, err
				}
				if len(stockPriceList) != 2 {
					return nil, nil
				}
				curStockPrice := stockPriceList[0]
				preStockPrice := stockPriceList[1]
				if preStockPrice.Amount == 0 {
					return nil, nil
				}
				if float64(curStockPrice.Amount)/float64(preStockPrice.Amount) < VolumeReportDiffThreshold {
					return nil, nil
				}
				// 如果是放量下跌则返回为空
				if curStockPrice.PriceClose < preStockPrice.PriceClose {
					return nil, nil
				}
				report := &model.VolumeReportItem{
					Code:          stockCode.CompanyCode,
					Name:          stockCode.CompanyName,
					PreDate:       utils.FormatDate(preStockPrice.Date),
					CurrentDate:   utils.FormatDate(curStockPrice.Date),
					PreAmount:     preStockPrice.Amount,
					CurrentAmount: curStockPrice.Amount,
					Diff:          utils.Float64KeepDecimal(float64(curStockPrice.Amount)/float64(preStockPrice.Amount), 2),
				}
				return report, nil
			}
		}(stockCode))
	}
	// 执行并发任务
	reportList, err := utils.ConcurrentActuator(jobList, MaxVolumeReportJobNum)
	if err != nil {
		return nil, err
	}
	var ret []*model.VolumeReportItem
	for _, item := range reportList {
		if item != nil {
			ret = append(ret, item.(*model.VolumeReportItem))
		}
	}
	// 对结果进行排序
	sort.Sort(model.VolumeReportItemSorter(ret))
	return ret, nil
}

func BuildPriceAnalyseReportMessage(result *model.PriceAnalyseReport) *model.LarkMessage {
	tableElement := model.TableElement{
		Tag:       "table",
		RowHeight: "middle",
		HeaderStyle: model.HeaderStyle{
			BackgroundStyle: "none",
			Bold:            true,
			Lines:           1,
		},
		Margin:   "0px 0px 0px 0px",
		PageSize: len(result.Items),
		Columns: []model.Column{
			{
				DataType:        "text",
				Name:            "name",
				DisplayName:     "股票名称",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "is_safe",
				DisplayName:     "最新状态",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "number",
				Name:            "count",
				DisplayName:     "连续天数",
				HorizontalAlign: "left",
				Width:           "auto",
			},
		},
		Rows: make([]map[string]interface{}, 0),
	}
	for _, item := range result.Items {
		tableElement.Rows = append(tableElement.Rows, map[string]interface{}{
			"name":    item.Name,
			"is_safe": item.IsSafe,
			"count":   item.Count,
		})
	}

	message := &model.LarkMessage{
		MsgType: "interactive",
		Card: model.LarkCard{
			Header: model.LarkHeader{
				Title: model.LarkTitle{
					Tag:     "plain_text",
					Content: "股票量价关系分析总结",
				},
				Subtitle: model.LarkTitle{
					Tag:     "plain_text",
					Content: fmt.Sprintf("个股截止%s的量价关系分析结论", result.EndDate),
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
						Content:   fmt.Sprintf("根据个股近%d天的量价关系数据, 着重观察持续的天数, 如果持续天数为1表示状态转折, 如果持续天数较长, 则可以高优先级看下, 可能形成了趋势.", AnalyzeVolumePriceLimit),
						TextAlign: "left",
						TextSize:  "normal_v2",
						Margin:    "0px 0px 0px 0px",
					},
				},
			},
		},
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, tableElement)
	return message
}

func BuildSubscribeStrategyReportMessage(data []*model.SubscribeStrategyResult) *model.LarkMessage {
	tableElement := model.TableElement{
		Tag:       "table",
		RowHeight: "middle",
		HeaderStyle: model.HeaderStyle{
			BackgroundStyle: "none",
			Bold:            true,
			Lines:           1,
		},
		Margin:   "0px 0px 0px 0px",
		PageSize: len(data),
		Columns: []model.Column{
			{
				DataType:        "text",
				Name:            "name",
				DisplayName:     "股票名称",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "strategy_detail",
				DisplayName:     "策略名称",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "number",
				Name:            "count",
				DisplayName:     "符合连续天数",
				HorizontalAlign: "left",
				Width:           "auto",
			},
			{
				DataType:        "text",
				Name:            "strategy",
				DisplayName:     "策略当前分析结果",
				HorizontalAlign: "left",
				Width:           "auto",
			},
		},
		Rows: make([]map[string]interface{}, 0),
	}
	for _, item := range data {
		tableElement.Rows = append(tableElement.Rows, map[string]interface{}{
			"name":            item.Code,
			"strategy_detail": item.StrategyDetail,
			"count":           item.Count,
			"strategy":        item.Strategy,
		})
	}

	message := &model.LarkMessage{
		MsgType: "interactive",
		Card: model.LarkCard{
			Header: model.LarkHeader{
				Title: model.LarkTitle{
					Tag:     "plain_text",
					Content: "股票订阅分析通知",
				},
				Subtitle: model.LarkTitle{
					Tag:     "plain_text",
					Content: fmt.Sprintf("根据您订阅的策略, 我们根据最新的股价进行了分析"),
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
						Content:   "根据订阅的策略, 只通知符合策略的股票信息, 并展示符合策略的连续天数, 期望您根据订阅的策略来做出相应的操作, 如果不再需要改策略, 请删除改策略以减少通知次数",
						TextAlign: "left",
						TextSize:  "normal_v2",
						Margin:    "0px 0px 0px 0px",
					},
				},
			},
		},
	}
	message.Card.Body.Elements = append(message.Card.Body.Elements, tableElement)
	return message
}
