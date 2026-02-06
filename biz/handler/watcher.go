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

func AddWatcher(ctx context.Context, c *app.RequestContext) {
	var req model.AddWatcherReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.AddWatcher(ctx, &req)
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

func GetWatchers(ctx context.Context, c *app.RequestContext) {
	var req model.GetWatchersReq
	if err := c.BindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	watchers, err := service.GetWatchers(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, utils.H{
		"message": "success",
		"data":    watchers,
	})
}

func DeleteWatcher(ctx context.Context, c *app.RequestContext) {
	var req model.DeleteWatcherReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.DeleteWatcher(ctx, &req)
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
