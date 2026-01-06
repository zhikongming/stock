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

func AddSubscribeStrategyData(ctx context.Context, c *app.RequestContext) {
	var strategy model.AddSubscribeStrategyReq
	if err := c.BindJSON(&strategy); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.AddSubscribeStrategyData(ctx, &strategy)
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

func GetSubscribeStrategyData(ctx context.Context, c *app.RequestContext) {
	var strategy model.GetSubscribeStrategyReq
	if err := c.BindQuery(&strategy); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	subscribeList, err := service.GetSubscribeStrategyData(ctx, &strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, utils.H{
		"message": "success",
		"data":    subscribeList,
	})
}

func DeleteSubscribeStrategyData(ctx context.Context, c *app.RequestContext) {
	var strategy model.DeleteSubscribeStrategyReq
	if err := c.BindJSON(&strategy); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}
	err := service.DeleteSubscribeStrategyData(ctx, &strategy)
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
