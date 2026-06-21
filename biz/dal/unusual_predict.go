package dal

import (
	"context"

	"gorm.io/gorm"
)

// UnusualPredict 严重异动预测数据库模型
type UnusualPredict struct {
	ID            uint    `gorm:"primaryKey"`
	Date          string  `gorm:"column:date"`
	Code          string  `gorm:"column:code"`
	Name          string  `gorm:"column:name"`
	PredictType   int     `gorm:"column:predict_type"`
	ChangeRate    float64 `gorm:"column:change_rate"`
	DeviationDay  int     `gorm:"column:deviation_day"`
	DeviationRate float64 `gorm:"column:deviation_rate"`
	PredictRate   float64 `gorm:"column:predict_rate"`
	RuleType      int     `gorm:"column:rule_type"`
}

func (u *UnusualPredict) TableName() string {
	return "unusual_predict"
}

// CreateUnusualPredict 创建异动预测记录
func CreateUnusualPredict(ctx context.Context, predict *UnusualPredict) error {
	db := GetDB()
	result := db.WithContext(ctx).Create(predict)
	return result.Error
}

// GetLastUnusualPredict 获取最新的异动预测列表
func GetLastUnusualPredict(ctx context.Context) (*UnusualPredict, error) {
	db := GetDB()
	var predict UnusualPredict
	err := db.WithContext(ctx).Order("date DESC").First(&predict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &predict, nil
}

func GetUnusualPredictByDateTypeRule(ctx context.Context, code string, date string, predictType int, ruleType int) (*UnusualPredict, error) {
	db := GetDB()
	var predict UnusualPredict
	err := db.WithContext(ctx).Where("code = ? AND date = ? AND predict_type = ? AND rule_type = ?", code, date, predictType, ruleType).First(&predict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &predict, nil
}

// GetUnusualPredictByDate 获取指定日期的异动预测
func GetUnusualPredictByDate(ctx context.Context, date string) ([]*UnusualPredict, error) {
	db := GetDB()
	var predicts []*UnusualPredict
	result := db.WithContext(ctx).Where("date = ?", date).Find(&predicts)
	return predicts, result.Error
}

// DeleteUnusualPredictByDate 删除指定日期的异动预测
func DeleteUnusualPredictByDate(ctx context.Context, date string) error {
	db := GetDB()
	result := db.WithContext(ctx).Where("date = ?", date).Delete(&UnusualPredict{})
	return result.Error
}
