package dal

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	StockTypeEastmoney = 0
)

type Watcher struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"column:name"`
	Stocks     string    `json:"stocks" gorm:"column:stocks"`
	StockType  int       `json:"stock_type" gorm:"column:stock_type"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`
	Status     int       `json:"status" gorm:"column:status"`
}

func (Watcher) TableName() string {
	return "watcher"
}

func CreateWatcher(ctx context.Context, watcher *Watcher) error {
	db := GetDB()
	return db.WithContext(ctx).Create(watcher).Error
}

func GetWatcher(ctx context.Context, id uint) (*Watcher, error) {
	db := GetDB()
	var watcher Watcher
	err := db.WithContext(ctx).Where("id = ? and status = ?", id, StatusEnabled).First(&watcher).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, nil
	}
	return &watcher, nil
}

func GetWatchers(ctx context.Context) ([]*Watcher, error) {
	db := GetDB()
	var watchers []*Watcher
	err := db.WithContext(ctx).Where("status = ?", StatusEnabled).Find(&watchers).Error
	if err != nil {
		return nil, err
	}
	return watchers, nil
}

func DeleteWatcher(ctx context.Context, id uint) error {
	db := GetDB()
	err := db.WithContext(ctx).Model(&Watcher{}).Where("id = ?", id).Update("status", StatusDisabled).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateWatcher(ctx context.Context, watcher *Watcher) error {
	db := GetDB()
	return db.WithContext(ctx).Save(watcher).Error
}
