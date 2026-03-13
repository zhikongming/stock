package dal

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type CacheType int8
type KeyType string

const (
	CacheTypeSimilarCompany   CacheType = 0
	CacheTypeVolumePrice      CacheType = 1
	CacheTypeScoreResult      CacheType = 2
	CacheTypeBusinessAnalysis CacheType = 3

	CacheKeyScoreResult KeyType = "score_result"
)

type Cache struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	DataKey   string `json:"data_key" gorm:"column:data_key"`
	DataType  int8   `json:"data_type" gorm:"column:data_type"`
	Date      string `json:"date" gorm:"column:date"`
	DataValue string `json:"data_value" gorm:"column:data_value"`
}

func (Cache) TableName() string {
	return "cache"
}

func GetCacheByType(ctx context.Context, key string, cacheType CacheType) (*Cache, error) {
	db := GetDB()
	var cache Cache
	err := db.WithContext(ctx).Where("data_key = ? and data_type = ?", key, cacheType).First(&cache).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cache, nil
}

func GetCacheByTypeDate(ctx context.Context, key string, cacheType CacheType, date string) (*Cache, error) {
	db := GetDB()
	var cache Cache
	err := db.WithContext(ctx).Where("data_key = ? and data_type = ? and date = ?", key, cacheType, date).Order("id desc").First(&cache).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cache, nil
}

func CreateCache(ctx context.Context, cache *Cache) error {
	db := GetDB()
	return db.WithContext(ctx).Create(cache).Error
}

func DeleteCache(ctx context.Context, id uint64) error {
	db := GetDB()
	return db.WithContext(ctx).Delete(&Cache{}, id).Error
}
