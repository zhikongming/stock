package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/zhikongming/stock/biz/service"
)

// CreateUnusualPredict 创建异动预测记录
func CreateUnusualPredict(ctx context.Context, c *app.RequestContext) {
	err := service.CreateUnusualPredict(ctx)
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

// GetUnusualPredictList 获取异动预测列表
func GetUnusualPredictList(ctx context.Context, c *app.RequestContext) {
	resp, err := service.GetUnusualPredictList(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"message": fmt.Sprintf("error: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
