package model

type PredictType int

const (
	PredictTypeNoUnusual      PredictType = iota // 0: 当日未异动
	PredictTypeTodayUnusual                      // 1: 当日异动
	PredictTypeNextDayUnusual                    // 2: 次日预测异动
)

var (
	PredictRuleTypeMap = map[int]string{
		1: "主板连续10个交易日内4次出现同向异常波动",
		3: "科创板连续10个交易日内3次出现同向异常波动",
		4: "连续十个交易日内日收盘价涨跌幅偏离值累计达到+100%",
		5: "连续十个交易日内日收盘价涨跌幅偏离值累计达到-50%",
		6: "连续三十个交易日内日收盘价涨跌幅偏离值累计达到+200%",
		7: "连续三十个交易日内日收盘价涨跌幅偏离值累计达到-70%",
	}
)

type UnusualPredict struct {
	Date          string      `json:"date"`           // 预测日期
	Code          string      `json:"code"`           // 股票代码
	Name          string      `json:"name"`           // 股票名称
	PredictType   PredictType `json:"predict_type"`   // 预测类型
	ChangeRate    float64     `json:"change_rate"`    // 当日涨跌幅
	DeviationDay  int         `json:"deviation_day"`  // 偏离天数
	DeviationRate float64     `json:"deviation_rate"` // 偏离涨跌幅
	PredictRate   float64     `json:"predict_rate"`   // 预测涨跌幅
	RuleType      int         `json:"rule_type"`      // 规则类型
	Rule          string      `json:"rule"`           // 规则描述
}

type PredictData struct {
	PredictType PredictType `json:"predict_type"` // 预测类型
	PredictRate float64     `json:"predict_rate"` // 预测涨跌幅
}

type MergedUnusualPredict struct {
	Date           string       `json:"date"`             // 预测日期
	Code           string       `json:"code"`             // 股票代码
	Name           string       `json:"name"`             // 股票名称
	TodayPredict   *PredictData `json:"today_predict"`    // 当日预测数据
	NextDayPredict *PredictData `json:"next_day_predict"` // 次日预测数据
	ChangeRate     float64      `json:"change_rate"`      // 当日涨跌幅
	DeviationDay   int          `json:"deviation_day"`    // 偏离天数
	DeviationRate  float64      `json:"deviation_rate"`   // 偏离涨跌幅
	RuleType       int          `json:"rule_type"`        // 规则类型
	Rule           string       `json:"rule"`             // 规则描述
}

type EMUnusualPredictResp struct {
	Result int    `json:"result"`
	Msg    string `json:"msg"`
	Pages  int    `json:"pages"`
	Date   int    `json:"date"`
	Open   int    `json:"open"`
	Count  int    `json:"count"`
	Data   []struct {
		M int     `json:"m"`
		C string  `json:"c"`
		N string  `json:"n"`
		S int     `json:"s"`
		E int     `json:"e"`
		X float64 `json:"x"`
		D int     `json:"d"`
		T float64 `json:"t"`
		A float64 `json:"a"`
		O int     `json:"o"`
	} `json:"data"`
}

func GetUnusualPredictRuleDesc(ruleType int) string {
	return PredictRuleTypeMap[ruleType]
}
