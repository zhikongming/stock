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
	EastMoneyIndustryPath      = "/api/qt/clist/get"
	EastMoneyFundFlowPath      = "/api/qt/stock/fflow/daykline/get"

	KLineTypeDay   = "101"
	KLineType30Min = "30"
)

var (
	NidList = []string{
		"nid18=04f8a68737c7fe9b73c36e870f1788eb;",
	}
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

func (c *EastMoneyClient) GetFullStockCode(code string) string {
	for matchPrefix, codePrefix := range utils.CodeToPrefixMap {
		if strings.HasPrefix(code, matchPrefix) {
			return fmt.Sprintf("%s%s", codePrefix, code)
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
	return c.GetRemoteStockBasic(ctx, code, dateTime, KLineTypeDay)
}

func (c *EastMoneyClient) GetRemoteStockBasic(ctx context.Context, code string, dateTime time.Time, kLintType string) (*model.StockDailyData, error) {
	path := fmt.Sprintf("%s%s", EastMoneyDomain3, EastMoneyStockDailyPath)
	params := map[string]string{
		"secid":   c.GetEastMoneyId(code),
		"end":     utils.FormatDate2(dateTime),
		"fields1": "f1,f2,f3,f4,f5,f6",
		"fields2": "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61",
		"klt":     kLintType,
		"fqt":     "1",
		"lmt":     "30",
	}
	var resp []byte
	var err error
	for i := 0; i < len(NidList); i++ {
		idx := emCache.GetCookieIndex()
		headers := map[string]string{
			"Cookie": NidList[idx],
		}
		resp, err = DoGet(ctx, path, params, headers)
		if err == nil {
			break
		}
		emCache.SetCookieIndex(idx + 1)
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
		timestamp := ParseTimeByKLineType(itemList[0], kLintType)
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

func ParseTimeByKLineType(date string, kLineType string) time.Time {
	switch kLineType {
	case KLineType30Min:
		return utils.ParseShortTime(date)
	}
	return utils.ParseDate(date)
}

func (c *EastMoneyClient) GetRemoteStockByKLineType(ctx context.Context, code string, startTime time.Time, endTime time.Time, kLineType model.KLineType) (*model.StockDailyData, error) {
	var localKLineType string
	switch kLineType {
	case model.KLineTypeDay:
		localKLineType = KLineTypeDay
	case model.KLineType30Min:
		localKLineType = KLineType30Min
	}
	return c.GetRemoteStockBasic(ctx, code, endTime, localKLineType)
}

func (c *EastMoneyClient) GetRemoteStockIndustry(ctx context.Context) ([]*model.IndustryItem, error) {
	path := fmt.Sprintf("%s%s", EastMoneyDomain2, EastMoneyIndustryPath)
	params := map[string]string{
		"fs":     "m:90+t:2+f:!50",
		"fields": "f12,f14",
		"fid":    "f13",
		"pn":     "1",
		"pz":     "200",
	}
	headers := map[string]string{}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.EMGetRemoteStockIndustryResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}

	data := make([]*model.IndustryItem, 0)
	if ret.Data != nil && ret.Data.Total > 0 {
		for _, item := range ret.Data.Diff {
			d := &model.IndustryItem{}
			if code, ok := item["f12"]; ok {
				d.Code = fmt.Sprintf("%v", code)
			}
			if name, ok := item["f14"]; ok {
				d.Name = fmt.Sprintf("%v", name)
			}
			data = append(data, d)
		}
	}

	return data, nil
}

func (c *EastMoneyClient) GetRemoteStockIndustryDetail(ctx context.Context, code string) ([]*model.StockItem, error) {
	path := fmt.Sprintf("%s%s", EastMoneyDomain2, EastMoneyIndustryPath)
	params := map[string]string{
		"fs":     fmt.Sprintf("b:%s", code),
		"fields": "f12,f14",
		"pn":     "1",
		"pz":     "1000",
	}
	headers := map[string]string{}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.EMGetRemoteStockIndustryResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}

	data := make([]*model.StockItem, 0)
	if ret.Data != nil && ret.Data.Total > 0 {
		for _, item := range ret.Data.Diff {
			d := &model.StockItem{}
			if code, ok := item["f12"]; ok {
				codeStr := fmt.Sprintf("%v", code)
				// 200开头的属于港股，忽略板块走势内
				if strings.HasPrefix(codeStr, utils.IgnoreCode) {
					continue
				}
				d.Code = c.GetFullStockCode(codeStr)
			}
			if name, ok := item["f14"]; ok {
				d.Name = fmt.Sprintf("%v", name)
			}
			data = append(data, d)
		}
	}

	return data, nil
}

func (c *EastMoneyClient) GetLatestRemoteFundFlow(ctx context.Context) ([]*model.FundFlowData, error) {
	data := make([]*model.FundFlowData, 0)

	path := fmt.Sprintf("%s%s", EastMoneyDomain2, EastMoneyIndustryPath)
	pageSize := 100
	params := map[string]string{
		"fid":    "f62",
		"po":     "1",
		"pz":     fmt.Sprintf("%d", pageSize),
		"np":     "1",
		"fltt":   "2",
		"invt":   "2",
		"fs":     "m:0+t:6+f:!2,m:0+t:13+f:!2,m:0+t:80+f:!2,m:1+t:2+f:!2,m:1+t:23+f:!2,m:0+t:7+f:!2,m:1+t:3+f:!2",
		"fields": "f12,f14,f2,f3,f62,f184,f66,f69,f72,f75,f78,f81,f84,f87,f204,f205,f124,f1,f13",
	}
	pageOffset := 1
	for {
		params["pn"] = fmt.Sprintf("%d", pageOffset)
		resp, err := DoGet(ctx, path, params, nil)
		if err != nil {
			return nil, err
		}

		var ret model.EMGetRemoteFundFlowResp
		err = json.Unmarshal(resp, &ret)
		if err != nil {
			log.Printf("json unmarshal failed: %v", err)
			return nil, err
		}

		if ret.Data == nil || ret.Data.Total <= 0 {
			break
		}
		for _, item := range ret.Data.Diff {
			d := &model.FundFlowData{}
			if code, ok := item["f12"]; ok {
				d.Code = c.GetFullStockCode(fmt.Sprintf("%v", code))
			}
			if name, ok := item["f14"]; ok {
				d.Name = fmt.Sprintf("%v", name)
			}
			if mainInflowAmount, ok := item["f62"]; ok {
				d.MainInflowAmount = int64(utils.ToFloat64(mainInflowAmount))
			}
			if extremeLargeInflowAmount, ok := item["f66"]; ok {
				d.ExtremeLargeInflowAmount = int64(utils.ToFloat64(extremeLargeInflowAmount))
			}
			if largeInflowAmount, ok := item["f72"]; ok {
				d.LargeInflowAmount = int64(utils.ToFloat64(largeInflowAmount))
			}
			if mediumInflowAmount, ok := item["f78"]; ok {
				d.MediumInflowAmount = int64(utils.ToFloat64(mediumInflowAmount))
			}
			if smallInflowAmount, ok := item["f84"]; ok {
				d.SmallInflowAmount = int64(utils.ToFloat64(smallInflowAmount))
			}
			if priceClose, ok := item["f2"]; ok {
				d.PriceClose = utils.ToFloat64(priceClose)
			}
			data = append(data, d)
		}

		if ret.Data.Total <= pageSize*pageOffset {
			break
		}

		pageOffset++
	}

	return data, nil
}

func (c *EastMoneyClient) GetRemoteFundFlowByCode(ctx context.Context, code string) ([]*model.FundFlowData, error) {
	data := make([]*model.FundFlowData, 0)

	path := fmt.Sprintf("%s%s", EastMoneyDomain3, EastMoneyFundFlowPath)
	params := map[string]string{
		"lmt":     "0",
		"klt":     "101",
		"fields1": "f1,f2,f3,f7",
		"fields2": "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61,f62,f63,f64,f65",
		"secid":   c.GetEastMoneyId(code),
	}
	headers := map[string]string{}
	resp, err := DoGet(ctx, path, params, headers)
	if err != nil {
		return nil, err
	}

	var ret model.EMGetRemoteDailyFundFlowResp
	err = json.Unmarshal(resp, &ret)
	if err != nil {
		log.Printf("json unmarshal failed: %v", err)
		return nil, err
	}

	if ret.Data == nil {
		return nil, fmt.Errorf("data is nil of %s", code)
	}

	for _, item := range ret.Data.Klines {
		itemList := strings.Split(item, ",")
		if len(itemList) < 14 {
			continue
		}
		data = append(data, &model.FundFlowData{
			Code:                     ret.Data.Code,
			Name:                     ret.Data.Name,
			MainInflowAmount:         int64(utils.ToFloat64(itemList[1])),
			ExtremeLargeInflowAmount: int64(utils.ToFloat64(itemList[5])),
			LargeInflowAmount:        int64(utils.ToFloat64(itemList[4])),
			MediumInflowAmount:       int64(utils.ToFloat64(itemList[3])),
			SmallInflowAmount:        int64(utils.ToFloat64(itemList[2])),
			PriceClose:               utils.ToFloat64(itemList[11]),
			Date:                     itemList[0],
		})
	}

	return data, nil
}
