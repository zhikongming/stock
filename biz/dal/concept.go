package dal

import (
	"context"

	"gorm.io/gorm"
)

type Concept struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"column:name"`
	Stocks string `gorm:"column:stocks"`
}

func (c *Concept) TableName() string {
	return "concept"
}

// GetConcepts 获取所有概念
func GetConcepts(ctx context.Context) ([]*Concept, error) {
	var concepts []*Concept
	result := db.WithContext(ctx).Find(&concepts)
	return concepts, result.Error
}

// GetConcept 根据ID获取概念
func GetConcept(ctx context.Context, id uint) (*Concept, error) {
	var concept Concept
	result := db.WithContext(ctx).First(&concept, id)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &concept, result.Error
}

// CreateConcept 创建概念
func CreateConcept(ctx context.Context, concept *Concept) error {
	result := db.WithContext(ctx).Create(concept)
	return result.Error
}

// UpdateConcept 更新概念
func UpdateConcept(ctx context.Context, concept *Concept) error {
	result := db.WithContext(ctx).Save(concept)
	return result.Error
}

// DeleteConcept 删除概念
func DeleteConcept(ctx context.Context, id uint) error {
	result := db.WithContext(ctx).Delete(&Concept{}, id)
	return result.Error
}

// IsConceptExist 检查概念是否存在
func IsConceptExist(ctx context.Context, name string) (bool, error) {
	var count int64
	result := db.WithContext(ctx).Model(&Concept{}).Where("name = ?", name).Count(&count)
	return count > 0, result.Error
}
