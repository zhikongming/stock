package service

import (
	"context"
	"time"

	"github.com/zhikongming/stock/biz/model"
)

type RemoteClient interface {
	GetRemoteStockCode(ctx context.Context, code string) (*model.StockBasicDataCompany, error)
	GetRemoteStockRelation(ctx context.Context, code string) ([]*model.StockRelationItem, error)
	GetRemoteStockDaily(ctx context.Context, code string, dateTime time.Time) (*model.StockDailyData, error)
	GetRemoteStockByKLineType(ctx context.Context, code string, startTime time.Time, endTime time.Time, kLineType model.KLineType) (*model.StockDailyData, error)
}
