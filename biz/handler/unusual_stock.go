package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/zhikongming/stock/biz/service"
)

// CreateUnusualStock 创建异常股票记录
func CreateUnusualStock(ctx context.Context, c *app.RequestContext) {
	err := service.CreateUnusualStock(ctx)
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

// GetUnusualStockList 获取异常股票列表
func GetUnusualStockList(ctx context.Context, c *app.RequestContext) {
	resp, err := service.GetUnusualStockList(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
