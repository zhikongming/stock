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
		"message": "pong",
	})
}
