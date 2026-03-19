package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/biz/service"
)

// AddPriceAnalyse 添加量价分析股票
func AddPriceAnalyse(ctx context.Context, c *app.RequestContext) {
	var req model.UpdatePriceAnalyseReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.UpdatePriceAnalyse(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("%v", err),
		})
		return
	}
	c.JSON(http.StatusOK, utils.H{
		"message": "success",
	})
}

// GetPriceAnalyse 获取量价分析结果
func GetPriceAnalyse(ctx context.Context, c *app.RequestContext) {
	var req model.GetPriceAnalyseReq
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	results, err := service.GetPriceAnalyse(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, utils.H{
		"message": "success",
		"data":    results,
	})
}
