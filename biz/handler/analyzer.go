package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/biz/service"
)

func AnalyzeStockCode(ctx context.Context, c *app.RequestContext) {
	var req model.AnalyzeStockCodeReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	var data *model.AnalyzeStockCodeResp
	var err error
	switch req.Strategy {
	case model.StockStrategyMa:
		data, err = service.AnalyzeMa(ctx, req)
	case model.StockStrategyBolling:
		data, err = service.AnalyzeBolling(ctx, req)
	case model.StockStrategyMacd:
		data, err = service.AnalyzeMacd(ctx, req)
	case model.StockStrategyKdj:
		data, err = service.AnalyzeKdj(ctx, req)
	default:
		result := make(map[string]*model.AnalyzeStockCodeResp)
		result["ma"], err = service.AnalyzeMa(ctx, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{
				"message": fmt.Sprintf("internal server error: %v", err),
			})
			return
		}
		result["bolling"], err = service.AnalyzeBolling(ctx, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{
				"message": fmt.Sprintf("internal server error: %v", err),
			})
			return
		}
		result["macd"], err = service.AnalyzeMacd(ctx, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{
				"message": fmt.Sprintf("internal server error: %v", err),
			})
			return
		}
		result["kdj"], err = service.AnalyzeKdj(ctx, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.H{
				"message": fmt.Sprintf("internal server error: %v", err),
			})
			return
		}
		c.JSON(consts.StatusOK, result)
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("internal server error: %v", err),
		})
		return
	}
	c.JSON(consts.StatusOK, data)
	return
}

func FilterStockCode(ctx context.Context, c *app.RequestContext) {
	var req model.FilterStockCodeReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	data, err := service.FilterStockCode(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("internal server error: %v", err),
		})
		return
	}
	c.JSON(consts.StatusOK, data)
	return
}

func AnalyzeTrendCode(ctx context.Context, c *app.RequestContext) {
	var req model.AnalyzeTrendCodeReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	if req.Code == "" || req.StartDate == "" {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	data, err := service.AnalyzeTrendCode(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("internal server error: %v", err),
		})
		return
	}
	c.JSON(consts.StatusOK, data)
	return
}
