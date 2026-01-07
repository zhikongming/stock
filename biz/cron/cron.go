package cron

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/robfig/cron/v3"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/biz/service"
)

func InitCron() {
	ctx := context.Background()
	c := cron.New()

	// 每天下午同步股票价格数据, 规则: 分 时 日 月 周
	c.AddFunc("0 17 * * *", func() {
		req := &model.SyncStockCodeReq{}
		err := service.SyncStockCode(ctx, req)
		if err != nil {
			hlog.Errorf("SyncStockCode failed, err: %v", err)
		}
	})

	c.Start()
}
