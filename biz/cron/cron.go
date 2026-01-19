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
		// 同步板块数据
		req1 := &model.SyncStockIndustryReq{}
		err := service.SyncStockIndustry(ctx, req1)
		if err != nil {
			hlog.Errorf("SyncStockIndustry failed, err: %v", err)
		}

		// 同步股价数据
		req := &model.SyncStockCodeReq{}
		err = service.SyncStockCode(ctx, req)
		if err != nil {
			hlog.Errorf("SyncStockCode failed, err: %v", err)
		}

		// 同步资金流向数据
		req2 := &model.SyncFundFlowReq{}
		err = service.SyncFundFlow(ctx, req2)
		if err != nil {
			hlog.Errorf("SyncFundFlow failed, err: %v", err)
		}
	})

	c.Start()
}
