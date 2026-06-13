package dal

import (
	"context"

	"gorm.io/gorm"
)

type Event struct {
	ID      uint   `gorm:"primaryKey"`
	Date    string `gorm:"column:date"`
	Event   string `gorm:"column:event"`
	Comment string `gorm:"column:comment"`
	Stocks  string `gorm:"column:stocks"`
}

func (e *Event) TableName() string {
	return "event"
}

// GetEvents 获取所有事件，按日期降序排列
func GetEvents(ctx context.Context) ([]*Event, error) {
	var events []*Event
	result := db.WithContext(ctx).Order("date DESC").Find(&events)
	return events, result.Error
}

// GetEvent 根据ID获取事件
func GetEvent(ctx context.Context, id uint) (*Event, error) {
	var event Event
	result := db.WithContext(ctx).First(&event, id)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &event, result.Error
}

// CreateEvent 创建事件
func CreateEvent(ctx context.Context, event *Event) error {
	result := db.WithContext(ctx).Create(event)
	return result.Error
}

// UpdateEvent 更新事件
func UpdateEvent(ctx context.Context, event *Event) error {
	result := db.WithContext(ctx).Save(event)
	return result.Error
}

// DeleteEvent 删除事件
func DeleteEvent(ctx context.Context, id uint) error {
	result := db.WithContext(ctx).Delete(&Event{}, id)
	return result.Error
}
