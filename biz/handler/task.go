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

func SyncStockCode(ctx context.Context, c *app.RequestContext) {
	var req model.SyncStockCodeReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.SyncStockCode(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("internal server error: %v", err),
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message": "success",
	})
}

func GetAllCode(ctx context.Context, c *app.RequestContext) {
	codeList, err := service.GetAllCode(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("internal server error: %v", err),
		})
		return
	}
	c.JSON(consts.StatusOK, codeList)
}

func SyncStockIndustry(ctx context.Context, c *app.RequestContext) {
	var req model.SyncStockIndustryReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.SyncStockIndustry(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message": "success",
	})
}

func SyncFundFlow(ctx context.Context, c *app.RequestContext) {
	var req model.SyncFundFlowReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.SyncFundFlow(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"message": "success",
	})
}

func GetStockInfo(ctx context.Context, c *app.RequestContext) {
	var req model.GetStockInfoReq
	if c.BindQuery(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	info, err := service.GetStockInfo(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}

	c.JSON(consts.StatusOK, info)
}
