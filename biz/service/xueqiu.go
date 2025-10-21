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
	XueqiuToken  = "xq_a_token=9773bacc11404cb5ac8b0847c564eda3730e6b61;u=621744617617136; HMACCOUNT=A99AB6BBFAAA4C14; "
	XueqiuDomain = "https://stock.xueqiu.com"

	XueqiuStockBasicPath    = "/v5/stock/f10/cn/company.json"
	XueqiuStockDailyPath    = "/v5/stock/chart/kline.json"
	XueqiuStockRelationPath = "/v5/stock/bar/relation.json"
)

type XueqiuClient struct{}

func NewXueqiuClient() RemoteClient {
	return &XueqiuClient{}
}

func (c *XueqiuClient) GetRemoteStockCode(ctx context.Context, code string) (*model.StockBasicDataCompany, error) {
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

func (c *XueqiuClient) GetRemoteStockRelation(ctx context.Context, code string) ([]*model.StockRelationItem, error) {
	path := fmt.Sprintf("%s%s", XueqiuDomain, XueqiuStockRelationPath)
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

	var ret model.StockRelationResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}
	return ret.Data.StockItemList, nil
}

func (c *XueqiuClient) GetRemoteStockDaily(ctx context.Context, code string, dateTime time.Time) (*model.StockDailyData, error) {
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

func (c *XueqiuClient) GetRemoteStockByKLineType(ctx context.Context, code string, startTime time.Time, endTime time.Time, kLineType model.KLineType) (*model.StockDailyData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *XueqiuClient) GetRemoteStockIndustry(ctx context.Context) ([]*model.IndustryItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *XueqiuClient) GetRemoteStockIndustryDetail(ctx context.Context, code string) ([]*model.StockItem, error) {
	return nil, fmt.Errorf("not implemented")
}
