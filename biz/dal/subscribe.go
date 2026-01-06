package dal

import (
	"context"
	"time"
)

const (
	SubscribeStatusEnabled  = 1
	SubscribeStatusDisabled = 0
)

type Subscribe struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	DateTime time.Time `json:"date_time" gorm:"column:date_time"`
	Strategy string    `json:"strategy" gorm:"column:strategy"`
	Status   int       `json:"status" gorm:"column:status"`
}

func (Subscribe) TableName() string {
	return "subscribe"
}

func CreateSubscribe(ctx context.Context, subscribe *Subscribe) error {
	db := GetDB()
	err := db.WithContext(ctx).Create(subscribe).Error
	if err != nil {
		return err
	}
	return nil
}

func GetSubscribeById(ctx context.Context, id uint) (*Subscribe, error) {
	db := GetDB()
	var subscribe Subscribe
	err := db.WithContext(ctx).Where("id = ? AND status = ?", id, SubscribeStatusEnabled).First(&subscribe).Error
	if err != nil {
		return nil, err
	}
	return &subscribe, nil
}

func GetAllSubscribeList(ctx context.Context) ([]*Subscribe, error) {
	db := GetDB()
	var subscribeList []*Subscribe
	err := db.WithContext(ctx).Where("status = ?", SubscribeStatusEnabled).Find(&subscribeList).Error
	if err != nil {
		return nil, err
	}
	return subscribeList, nil
}

func DeleteSubscribeById(ctx context.Context, id uint) error {
	db := GetDB()
	err := db.WithContext(ctx).Model(&Subscribe{}).Where("id = ?", id).Update("status", SubscribeStatusDisabled).Error
	if err != nil {
		return err
	}
	return nil
}
