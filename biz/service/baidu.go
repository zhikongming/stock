package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

const (
	BaiduDomain          = "https://finance.pae.baidu.com"
	BaiduStockDailyPath  = "/vapi/v1/getquotation"
	BaiduShareholderPath = "/selfselect/openapi"
	BaiduCompanyInfoPath = "/api/stockwidget"
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

func (c *BaiduClient) GetCompanyCode(ctx context.Context, code string) (string, error) {
	stockCode, err := dal.GetStockCodeByCode(ctx, code)
	if err != nil {
		return "", err
	}
	if stockCode == nil {
		return "", fmt.Errorf("stock code not found")
	}
	// 如果没有百度公司代码，调用接口获取并缓存下来
	if stockCode.BdCompanyCode != "" {
		return stockCode.BdCompanyCode, nil
	}

	path := fmt.Sprintf("%s%s", BaiduDomain, BaiduCompanyInfoPath)
	code = utils.GetStockCodeNumber(code)
	params := map[string]string{
		"code":          code,
		"market":        "ab",
		"type":          "stock",
		"widgetType":    "company",
		"finClientType": "pc",
	}
	headers := map[string]string{
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
		return "", err
	}

	var ret model.GetRemoteCompanyInfoResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("code:%s, json unmarshal failed: %v", code, err)
		return "", err
	}

	// 缓存公司代码
	stockCode.BdCompanyCode = ret.Result.Content.NewCompany.BasicInfo.CompanyCode
	dal.UpdateStockCode(ctx, stockCode)

	return ret.Result.Content.NewCompany.BasicInfo.CompanyCode, nil
}

func (c *BaiduClient) GetRemoteShareholder(ctx context.Context, code string, date string) (*model.Top10Shareholder, error) {
	path := fmt.Sprintf("%s%s", BaiduDomain, BaiduShareholderPath)
	companyCode, err := c.GetCompanyCode(ctx, code)
	if err != nil {
		return nil, err
	}
	code = utils.GetStockCodeNumber(code)
	params := map[string]string{
		"srcid":         "5539",
		"code":          code,
		"company_code":  companyCode,
		"inner_code":    "1164",
		"group":         "holder_equity",
		"listedSector":  "1",
		"finClientType": "pc",
		"hold_date":     date,
	}
	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Accept":          "application/vnd.finance-web.v1+json",
		"Host":            "finance.pae.baidu.com",
		"Connection":      "keep-alive",
		"Accept-Encoding": "gzip, deflate, br, zstd",
		"Accept-Language": "zh-CN,zh;q=0.9",
		"origin":          "https://gushitong.baidu.com",
	}
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := DoGet(ctx2, path, params, headers)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			params["listedSector"] = "7"
			resp, err = DoGet(ctx, path, params, headers)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	var ret model.BDGetRemoteShareholderResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("code:%s, json unmarshal failed: %v", code, err)
		return nil, err
	}

	// 做数据格式的转换
	top10Shareholder := &model.Top10Shareholder{
		Num:             0,
		ShareholderList: make([]*model.Shareholder, 0),
	}
	tmpDate := strings.ReplaceAll(date, "-", "/")
	for _, item := range ret.Result.BDShareholders.BDShareholderList {
		if item.ReportDate == tmpDate {
			top10Shareholder.Num, _ = strconv.Atoi(item.NumOrigin)
		}
	}
	for _, item := range ret.Result.BDHoldShare.Content.Body {
		top10Shareholder.ShareholderList = append(top10Shareholder.ShareholderList, &model.Shareholder{
			ShareholderName:   item.Holder,
			ShareholderNumber: item.HoldNum,
			ShareholderPer:    item.HoldPer,
		})
	}
	return top10Shareholder, nil
}
