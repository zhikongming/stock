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

// CreateEvent 创建事件
func CreateEvent(ctx context.Context, c *app.RequestContext) {
	var req model.CreateEventReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.CreateEvent(ctx, &req)
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

// UpdateEvent 更新事件
func UpdateEvent(ctx context.Context, c *app.RequestContext) {
	var req model.UpdateEventReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.UpdateEvent(ctx, &req)
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

// DeleteEvent 删除事件
func DeleteEvent(ctx context.Context, c *app.RequestContext) {
	var req model.DeleteEventReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"message": "bad request",
		})
		return
	}

	err := service.DeleteEvent(ctx, &req)
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

// GetEventTimeline 获取事件时间轴
func GetEventTimeline(ctx context.Context, c *app.RequestContext) {
	timeline, err := service.GetEventTimeline(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("%v", err),
		})
		return
	}
	c.JSON(http.StatusOK, timeline)
}
