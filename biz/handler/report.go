package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/biz/service/report"
)

func GetStockReport(ctx context.Context, c *app.RequestContext) {
	var req model.GetStockReportReq
	err := c.BindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": err.Error(),
		})
		return
	}
	report, err := report.GetStockReport(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, report)
}

func AddStockReport(ctx context.Context, c *app.RequestContext) {
	var req model.AddStockReportReq
	if c.BindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	err := report.AddStockReport(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, utils.H{
		"message": "success",
	})
}

func GetBankTrackData(ctx context.Context, c *app.RequestContext) {
	var req model.GetBankTrackDataReq
	if c.BindQuery(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	data, err := report.GetBankTrackData(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("%v", err),
		})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetIndustryTrackData(ctx context.Context, c *app.RequestContext) {
	var req model.GetIndustryTrackDataReq
	if c.BindQuery(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	data, err := report.GetIndustryTrackData(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("%v", err),
		})
		return
	}
	c.JSON(http.StatusOK, data)
}
