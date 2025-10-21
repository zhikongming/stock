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

func GetIndustryBasicData(ctx context.Context, c *app.RequestContext) {
	var req model.GetIndustryBasicDataReq
	if c.BindQuery(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	data, err := service.GetIndustryBasicData(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, data)
}

// 获取板块的走势图
func GetIndustryTrendData(ctx context.Context, c *app.RequestContext) {
	var req model.GetIndustryTrendDataReq
	if c.BindQuery(&req) != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	if req.Days <= 0 {
		req.Days = 1
	} else if req.Days > 360 {
		req.Days = 360
	}
	data, err := service.GetIndustryTrendData(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, data)
}
