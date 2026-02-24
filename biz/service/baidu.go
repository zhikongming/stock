package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	BaiduDomain         = "https://finance.pae.baidu.com"
	BaiduStockDailyPath = "/vapi/v1/getquotation"
)

type BaiduClient struct {
}

func NewBaiduClient() RemoteClient {
	return &BaiduClient{}
}

func (c *BaiduClient) GetRemoteStockCode(ctx context.Context, code string) (*model.StockBasicDataCompany, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *BaiduClient) GetRemoteStockRelation(ctx context.Context, code string) ([]*model.StockRelationItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *BaiduClient) GetRemoteStockDaily(ctx context.Context, code string, dateTime time.Time) (*model.StockDailyData, error) {
	path := fmt.Sprintf("%s%s", BaiduDomain, BaiduStockDailyPath)
	code = utils.GetStockCodeNumber(code)
	params := map[string]string{
		"group":       "quotation_kline_ab",
		"market_type": "ab",
		"newFormat":   "1",
		"is_kc":       "0",
		"ktype":       "day",
		"query":       code,
		"code":        code,
	}
	headers := map[string]string{
		"Cookie":          "BAIDUID=5B502C05DF2CFBC60C7E28CCEC88A863:FG=1",
		"acs-token":       "1770962406840_1770994891027_0H8Hsi6BOiUvzDWJOZjxBbCefGKOp5A5n25L+PKxwWlpLHuGEhJpUIwQjBYS9ArAUTUVHBOGKRw1bN9Pr5/t5haKqCRZGEuFgh3osncDaHseWhJ8O6voJhvW/MsDxFZlUnkR1iQBGBCzq7QMIP4ba68hvkKrn22WY0lnQkvFlyyEEjHlANfzsEFGWwVSNVgJLFu/rEpRX6AcXEcYH3Jue3hzOgKFzScML2LRVhPwuwZl1O3SHL9zC+QA3EeWP2Whbzw9KbZ7GFw9Kty7AGgcM4sOXGetBmlyJugXlvm+rEwWfQ7PSwfO+TcTboYABHSdN6PDePeOFh7tn8VhWKxSx67UZpbl2PUlf1aqDD03Wh82V3GPqMZGj6k7SFDaJhFEkQ5LoxjEuCLn2MGnoADq4XODcJq0us6sW8rTqGa+uqkFoZKpO26kCYbXT+mIUDSKx7lWx10lgJ/4YznUM/UDcg==",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Accept":          "application/vnd.finance-web.v1+json",
		"Host":            "finance.pae.baidu.com",
		"Connection":      "keep-alive",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"Accept-Language": "zh-CN,zh;q=0.9",
		"origin":          "https://gushitong.baidu.com",
	}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.BDGetRemoteStockDailyResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("code:%s, json unmarshal failed: %v", code, err)
		return nil, err
	}

	data := &model.StockDailyData{
		Column: []string{
			"timestamp",
			"high",
			"low",
			"open",
			"close",
			"amount",
		},
		Item: make([][]interface{}, 0),
	}
	priceList := strings.Split(ret.Result.NewMarketData.MarketData, ";")
	timestampIndex := utils.Index("time", ret.Result.NewMarketData.Keys)
	highIndex := utils.Index("high", ret.Result.NewMarketData.Keys)
	lowIndex := utils.Index("low", ret.Result.NewMarketData.Keys)
	openIndex := utils.Index("open", ret.Result.NewMarketData.Keys)
	closeIndex := utils.Index("close", ret.Result.NewMarketData.Keys)
	amountIndex := utils.Index("amount", ret.Result.NewMarketData.Keys)
	if timestampIndex == -1 || highIndex == -1 || lowIndex == -1 || openIndex == -1 || closeIndex == -1 || amountIndex == -1 {
		return nil, fmt.Errorf("index not found")
	}
	for _, item := range priceList {
		itemList := strings.Split(item, ",")
		if len(itemList) < 7 {
			continue
		}
		open := itemList[openIndex]
		close := itemList[closeIndex]
		high := itemList[highIndex]
		low := itemList[lowIndex]
		amount := itemList[amountIndex]
		t := itemList[timestampIndex]
		timestamp := ParseTimeByKLineType(t, KLineTypeDay)
		data.Item = append(data.Item, []interface{}{
			utils.TimeToTimestamp(timestamp) * 1000,
			high,
			low,
			open,
			close,
			amount,
		})
	}

	return data, nil
}

func (c *BaiduClient) GetRemoteStockByKLineType(ctx context.Context, code string, startTime time.Time, endTime time.Time, kLineType model.KLineType) (*model.StockDailyData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *BaiduClient) GetRemoteStockIndustry(ctx context.Context) ([]*model.IndustryItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *BaiduClient) GetRemoteStockIndustryDetail(ctx context.Context, code string) ([]*model.StockItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *BaiduClient) GetLatestRemoteFundFlow(ctx context.Context) ([]*model.FundFlowData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *BaiduClient) GetRemoteFundFlowByCode(ctx context.Context, code string) ([]*model.FundFlowData, error) {
	return nil, fmt.Errorf("not implemented")
}
