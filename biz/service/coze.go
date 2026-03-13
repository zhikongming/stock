package service

import (
	"context"
	"encoding/json"

	"github.com/zhikongming/stock/biz/config"
	"github.com/zhikongming/stock/biz/dal"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/utils"
)

type CozeClient struct {
	GetSimilarCompanyUrl     string
	GetSimilarCompanyToken   string
	GetVolumePriceUrl        string
	GetVolumePriceToken      string
	GetBusinessAnalysisUrl   string
	GetBusinessAnalysisToken string
}

var cozeClient *CozeClient

func NewCozeClient() *CozeClient {
	if cozeClient == nil {
		cozeClient = &CozeClient{}
		conf := config.GetCozeConfig()
		if conf == nil {
			return nil
		}
		cozeClient.GetSimilarCompanyUrl = conf.GetSimilarCompanyUrl
		cozeClient.GetSimilarCompanyToken = conf.GetSimilarCompanyToken
		cozeClient.GetVolumePriceUrl = conf.GetVolumePriceUrl
		cozeClient.GetVolumePriceToken = conf.GetVolumePriceToken
		cozeClient.GetBusinessAnalysisUrl = conf.GetBusinessAnalysisUrl
		cozeClient.GetBusinessAnalysisToken = conf.GetBusinessAnalysisToken
	}
	return cozeClient
}

func (c *CozeClient) GetSimilarCompany(ctx context.Context, companyName string) ([]*model.SimilarCompany, error) {
	req := &model.GetSimilarCompanyReq{
		CompanyName: companyName,
	}
	resp, err := DoPost(ctx, c.GetSimilarCompanyUrl, nil, map[string]string{
		"Authorization": "Bearer " + c.GetSimilarCompanyToken,
		"Content-Type":  "application/json",
	}, req)
	if err != nil {
		return nil, err
	}
	var respBody model.GetSimilarCompanyResp
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		return nil, err
	}
	return respBody.SimilarCompanies, nil
}

func (c *CozeClient) GetVolumePrice(ctx context.Context, companyName string, stockPriceList []*dal.StockPrice) (*model.GetVolumePriceResp, error) {
	stockDataList := make([]*model.StockData, 0)
	for _, stockPrice := range stockPriceList {
		stockDataList = append(stockDataList, &model.StockData{
			Date:       utils.FormatDate(stockPrice.Date),
			OpenPrice:  utils.ToString(stockPrice.PriceOpen),
			ClosePrice: utils.ToString(stockPrice.PriceClose),
			HighPrice:  utils.ToString(stockPrice.PriceHigh),
			LowPrice:   utils.ToString(stockPrice.PriceLow),
			Volume:     utils.ToString(stockPrice.Amount),
		})
	}
	req := &model.GetVolumePriceReq{
		CompanyName:   companyName,
		StockDataList: stockDataList,
	}
	resp, err := DoPost(ctx, c.GetVolumePriceUrl, nil, map[string]string{
		"Authorization": "Bearer " + c.GetVolumePriceToken,
		"Content-Type":  "application/json",
	}, req)
	if err != nil {
		return nil, err
	}
	var respBody model.GetVolumePriceResp
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}

func (c *CozeClient) GetBusinessAnalysis(ctx context.Context, companyName string) (*model.GetBusinessAnalysisResp, error) {
	req := &model.GetBusinessAnalysisReq{
		CompanyName: companyName,
	}
	resp, err := DoPost(ctx, c.GetBusinessAnalysisUrl, nil, map[string]string{
		"Authorization": "Bearer " + c.GetBusinessAnalysisToken,
		"Content-Type":  "application/json",
	}, req)
	if err != nil {
		return nil, err
	}
	var respBody model.GetBusinessAnalysisResp
	err = json.Unmarshal(resp, &respBody)
	if err != nil {
		return nil, err
	}
	return &respBody, nil
}
