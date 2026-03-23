package cron

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/robfig/cron/v3"
	"github.com/zhikongming/stock/biz/model"
	"github.com/zhikongming/stock/biz/service"
	"github.com/zhikongming/stock/utils"
)

func InitCron() {
	ctx := context.Background()
	c := cron.New()

	// 每天下午同步股票价格数据, 规则: 分 时 日 月 周
	// 记得必须在下午三点后执行
	c.AddFunc("10 15 * * *", func() {
		// 如果是周六或者周日, 则不执行
		if utils.IsNowWeekend() {
			hlog.Infof("Today is weekend, skip sync stock price")
			return
		}
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

		// 计算报告数据
		service.GetAnalyzeReport(ctx)

		// 分析量价关系并发送报告
		service.GetPriceAnalyse(ctx, &model.GetPriceAnalyseReq{})
		service.GetPriceAnalyseReport(ctx)

		// 分析策略并发送报告
		service.GetSubscribeStrategyReport(ctx)
	})

	c.Start()
}
