package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zhikongming/stock/biz/model"
)

const (
	XueqiuToken  = "u=111714051563790; HMACCOUNT=5D0238B683BD8C9F; xq_a_token=220b0abef0fac476d076c9f7a3938b7edac35f48; "
	XueqiuDomain = "https://stock.xueqiu.com"

	XueqiuStockBasicPath = "/v5/stock/f10/cn/company.json"
	XueqiuStockDailyPath = "/v5/stock/chart/kline.json"
)

func GetRemoteStockBasic(ctx context.Context, code string) (*model.StockBasicDataCompany, error) {
	path := fmt.Sprintf("%s%s", XueqiuDomain, XueqiuStockBasicPath)
	params := map[string]string{
		"symbol": code,
	}
	headers := map[string]string{
		"Cookie": XueqiuToken,
	}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.GetRemoteStockBasicResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}
	return ret.Data.Company, nil
}

func GetRemoteStockDaily(ctx context.Context, code string, dateTime time.Time) (*model.StockDailyData, error) {
	path := fmt.Sprintf("%s%s", XueqiuDomain, XueqiuStockDailyPath)
	params := map[string]string{
		"symbol":    code,
		"begin":     fmt.Sprintf("%d", dateTime.UnixNano()/int64(time.Millisecond)),
		"period":    "day",
		"type":      "before",
		"count":     "-365",
		"indicator": "kline,pe,pb,ps,pcf,market_capital,agt,ggt,balance",
	}
	headers := map[string]string{
		"Cookie": XueqiuToken,
	}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}
	var ret model.GetRemoteStockDailyResp
	decoder := json.NewDecoder(strings.NewReader(string(resp)))
	decoder.UseNumber()
	err = decoder.Decode(&ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}
	return ret.Data, nil
}
