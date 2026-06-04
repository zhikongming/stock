package dal

import (
	"context"

	"gorm.io/gorm"
)

// UnusualStock 异常股票数据库模型
type UnusualStock struct {
	ID            uint   `gorm:"primaryKey"`
	Code          string `gorm:"column:code"`
	Name          string `gorm:"column:name"`
	Type          int    `gorm:"column:type"`
	StartDate     string `gorm:"column:start_date"`
	EndDate       string `gorm:"column:end_date"`
	NoticeDate    string `gorm:"column:notice_date"`
	UnusualType   string `gorm:"column:unusual_type"`
	UnusualReason string `gorm:"column:unusual_reason"`
}

func (u *UnusualStock) TableName() string {
	return "unusual_stock"
}

// CreateUnusualStock 创建异常股票记录
func CreateUnusualStock(ctx context.Context, stock *UnusualStock) error {
	db := GetDB()

	result := db.WithContext(ctx).Create(stock)
	return result.Error
}

// GetUnusualStockList 获取异常股票列表
func GetUnusualStockList(ctx context.Context) ([]*UnusualStock, error) {
	db := GetDB()
	var stocks []*UnusualStock

	result := db.WithContext(ctx).Find(&stocks)
	return stocks, result.Error
}

func GetUnusualStockByCodeTypeEndDate(ctx context.Context, code string, t int, endDate string) (*UnusualStock, error) {
	db := GetDB()
	var stock UnusualStock

	err := db.WithContext(ctx).Where("code = ? AND type = ? AND end_date = ?", code, t, endDate).First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stock, err
}
