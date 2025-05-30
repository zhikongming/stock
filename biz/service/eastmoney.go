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
	EastMoneyDomain  = "https://datacenter.eastmoney.com"
	EastMoneyDomain2 = "https://push2.eastmoney.com"
	EastMoneyDomain3 = "https://push2his.eastmoney.com"

	EastMoneyBasicPath         = "/securities/api/data/v1/get"
	EastMoneyStockRelationPath = "/api/qt/stock/get"
	EastMoneyStockDailyPath    = "/api/qt/stock/kline/get"
)

type EastMoneyClient struct {
}

func NewEastMoneyClient() RemoteClient {
	return &EastMoneyClient{}
}

func (c *EastMoneyClient) GetEastMoneyCode(code string) string {
	for _, pref := range utils.StockPrefixList {
		if strings.HasPrefix(code, pref) {
			return fmt.Sprintf("%s.%s", code[len(pref):], pref)
		}
	}
	return code
}

func (c *EastMoneyClient) GetEastMoneyId(code string) string {
	for prefName, prefCode := range utils.StockIdMap {
		if strings.HasPrefix(code, prefName) {
			return fmt.Sprintf("%s.%s", prefCode, code[len(prefName):])
		}
	}
	return code
}

func (c *EastMoneyClient) GetRemoteStockCode(ctx context.Context, code string) (*model.StockBasicDataCompany, error) {
	path := fmt.Sprintf("%s%s", EastMoneyDomain, EastMoneyBasicPath)
	params := map[string]string{
		"reportName": "RPT_F10_BASIC_ORGINFO",
		"columns":    "ALL",
		"filter":     fmt.Sprintf("(SECUCODE=\"%s\")", c.GetEastMoneyCode(code)),
	}
	headers := map[string]string{}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.EMGetRemoteStockBasicResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}

	if len(ret.Result.Data) > 0 {
		return &model.StockBasicDataCompany{
			OrgShortNameCN: ret.Result.Data[0].SecretaryNameAbbr,
			ClassiName:     "",
			ListedDate:     utils.TimeToTimestamp(utils.ParseTime(ret.Result.Data[0].ListingDate)) * 1000,
		}, nil
	}

	return nil, fmt.Errorf("stock code not found")
}

func (c *EastMoneyClient) GetRemoteStockRelation(ctx context.Context, code string) ([]*model.StockRelationItem, error) {
	path := fmt.Sprintf("%s%s", EastMoneyDomain2, EastMoneyStockRelationPath)
	params := map[string]string{
		"fields": "f57%2Cf58%2Cf256",
		"secid":  c.GetEastMoneyId(code),
	}
	headers := map[string]string{}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.EMStockRelationResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}

	hkcode := ""
	hkname := ""
	for k, v := range ret.Data {
		switch k {
		case "f256":
			valueStr := fmt.Sprintf("%v", v)
			if len(valueStr) > 0 {
				hkcode = valueStr
			}
		case "f58":
			valueStr := fmt.Sprintf("%v", v)
			if len(valueStr) > 0 {
				hkname = valueStr
			}
		}
	}
	if len(hkcode) > 0 && len(hkname) > 0 {
		return []*model.StockRelationItem{
			{
				Symbol: hkcode,
				Name:   hkname,
			},
		}, nil
	}

	return nil, nil
}

func (c *EastMoneyClient) GetRemoteStockDaily(ctx context.Context, code string, dateTime time.Time) (*model.StockDailyData, error) {
	path := fmt.Sprintf("%s%s", EastMoneyDomain3, EastMoneyStockDailyPath)
	params := map[string]string{
		"secid":   c.GetEastMoneyId(code),
		"end":     utils.FormatDate2(dateTime),
		"fields1": "f1,f2,f3,f4,f5,f6",
		"fields2": "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61",
		"klt":     "101",
		"fqt":     "1",
		"lmt":     "360",
	}
	headers := map[string]string{}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.EMGetRemoteStockDailyResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
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
	for _, item := range ret.Data.Klines {
		itemList := strings.Split(item, ",")
		if len(itemList) < 7 {
			continue
		}
		timestamp := utils.ParseDate(itemList[0])
		open := itemList[1]
		close := itemList[2]
		high := itemList[3]
		low := itemList[4]
		amount := itemList[6]
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
